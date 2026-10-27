package logfilesmetricexporter

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/prometheus"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

//go:embed valid.yaml
var validCR string

//go:embed invalid.yaml
var inValidCR string

var _ = Describe("[e2e][logfilemetricexporter] LogFileMetricsExporter", Serial, func() {

	defer GinkgoRecover()

	var (
		err        error
		e2e        = framework.NewE2ETestFramework()
		createLFME = func(cr string) error {
			lfme := &v1alpha1.LogFileMetricExporter{}
			test.MustUnmarshal(cr, lfme)
			return e2e.Create(lfme)
		}
	)
	AfterEach(func() {
		e2e.Cleanup()
	})

	It("should reject any CR not named openshift-logging/instance", func() {
		err = createLFME(inValidCR)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp("is invalid.*supported values.*instance"), "exp. the CR to be rejected because it is not THE singleton")
	})

	It("should serve metrics to authorized clients providing a valid bearer token", func() {
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc -n openshift-logging delete --ignore-not-found logfilemetricexporter instance").Output()
		})
		metricsReaderRoleName := fmt.Sprintf("%s-metrics-reader", constants.ClusterLoggingOperator)
		metricsReaderBindingName := fmt.Sprintf("%s-metrics-reader", constants.LogfilesmetricexporterName)
		metricsAuthRoleName := fmt.Sprintf("%s-metrics-auth", constants.LogfilesmetricexporterName)
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc delete --ignore-not-found clusterrole %s", metricsReaderRoleName).Output()
		})
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc delete --ignore-not-found clusterrolebinding %s", metricsReaderBindingName).Output()
		})
		// Delete the metrics auth ClusterRoleBinding
		// The LFME reconciles the ClusterRoleBinding and ClusterRole for metrics auth
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc delete --ignore-not-found clusterrolebinding %s", metricsAuthRoleName).Output()
		})

		err = createLFME(validCR)
		Expect(err).ToNot(HaveOccurred())
		Expect(e2e.WaitForDaemonSet(constants.OpenshiftNS, constants.LogfilesmetricexporterName)).To(Succeed())

		By("creating the metrics-reader ClusterRole")
		roleFilePath, err := filepath.Abs(filepath.Join("..", "..", "..", "config", "rbac", "metrics_reader_role.yaml"))
		Expect(err).ToNot(HaveOccurred(), "Failed to construct role file path")
		_, err = oc.Literal().From("oc apply -f %s", roleFilePath).Run()
		Expect(err).ToNot(HaveOccurred(), "Failed to create metrics-reader ClusterRole")

		By("creating a ClusterRoleBinding for the service account to allow access to metrics")
		_, err = oc.Literal().From("oc create clusterrolebinding %s --clusterrole=%s --serviceaccount=%s:%s",
			metricsReaderBindingName, metricsReaderRoleName, constants.OpenshiftNS, constants.LogfilesmetricexporterName).Run()
		Expect(err).ToNot(HaveOccurred(), "Failed to create ClusterRoleBinding")

		// Give the exporter time to start collecting logs
		time.Sleep(10 * time.Second)

		// Get a pod from the DaemonSet
		pods, err := e2e.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", constants.LabelK8sComponent, constants.LogfilesmetricexporterName),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pods.Items).ToNot(BeEmpty(), "Expected to find at least one LFME pod")

		podName := pods.Items[0].Name

		// Request a token for the LFME service account via the TokenRequest API
		tokenReq, err := e2e.KubeClient.CoreV1().ServiceAccounts(constants.OpenshiftNS).
			CreateToken(context.TODO(), constants.LogfilesmetricexporterName,
				&authv1.TokenRequest{}, metav1.CreateOptions{})
		Expect(err).ToNot(HaveOccurred(), "Failed to request service account token")
		token := tokenReq.Status.Token

		// Use port-forward to access the pod's metrics endpoint directly,
		// then make an HTTPS request with the bearer token for RBAC authentication
		transport, upgrader, err := spdy.RoundTripperFor(e2e.RestConfig)
		Expect(err).ToNot(HaveOccurred())

		pfURL := e2e.KubeClient.CoreV1().RESTClient().Post().
			Namespace(constants.OpenshiftNS).
			Resource("pods").
			Name(podName).
			SubResource("portforward").
			URL()

		dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, pfURL)

		stopChan := make(chan struct{}, 1)
		readyChan := make(chan struct{})
		defer close(stopChan)

		fw, err := portforward.New(dialer, []string{"0:2112"}, stopChan, readyChan, io.Discard, io.Discard)
		Expect(err).ToNot(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			Expect(fw.ForwardPorts()).To(Succeed())
		}()

		<-readyChan

		ports, err := fw.GetPorts()
		Expect(err).ToNot(HaveOccurred())
		Expect(ports).ToNot(BeEmpty())
		localPort := ports[0].Local

		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}

		metricsURL := fmt.Sprintf("https://localhost:%d/metrics", localPort)

		err = wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Second*30, true, func(context.Context) (done bool, err error) {
			req, err := http.NewRequest(http.MethodGet, metricsURL, nil)
			if err != nil {
				return false, nil
			}
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := httpClient.Do(req)
			if err != nil {
				log.V(5).Info("Failed to fetch secure metrics", "error", err)
				return false, nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return false, nil
			}
			out := string(body)
			log.V(5).Info("Polling secure metrics", "status", resp.StatusCode, "result", out)
			if resp.StatusCode != http.StatusOK {
				return false, nil
			}
			if !strings.Contains(out, "log_logged_bytes_total") {
				return false, nil
			}
			return regexp.MatchString(`log_logged_bytes_total{.*} [1-9][0-9]*`, out)
		})
		Expect(err).ToNot(HaveOccurred(), "Exp. to scrape metrics with a bearer token")
	})

	It("should have LFME metrics scraped by Prometheus via the ServiceMonitor", func() {
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc -n openshift-logging delete --ignore-not-found logfilemetricexporter instance").Output()
		})
		err = createLFME(validCR)
		Expect(err).ToNot(HaveOccurred())
		Expect(e2e.WaitForDaemonSet(constants.OpenshiftNS, constants.LogfilesmetricexporterName)).To(Succeed())

		By("querying Thanos for an LFME metric to validate the ServiceMonitor pipeline")
		Eventually(func(g Gomega) {
			response, err := prometheus.Query(`log_logged_bytes_total{namespace="openshift-logging"}`)
			g.Expect(err).NotTo(HaveOccurred(), "Failed to query metric")
			g.Expect(prometheus.HasResults(response)).To(BeTrue())
		}, 5*time.Minute, 30*time.Second).Should(Succeed(),
			"LFME metrics should appear in Prometheus, indicating the ServiceMonitor is correctly configured")
	})
})
