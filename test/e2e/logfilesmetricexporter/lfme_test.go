package logfilesmetricexporter

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

//go:embed valid.yaml
var validCR string

//go:embed invalid.yaml
var inValidCR string

var _ = Describe("[e2e][logfilemetricexporter] LogFileMetricsExporter", func() {

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

	It("should be deployed by the operator and producing metrics", func() {
		e2e.AddCleanup(func() error {
			return oc.Literal().From("oc -n openshift-logging delete --ignore-not-found logfilemetricexporter instance").Output()
		})
		err = createLFME(validCR)
		Expect(err).ToNot(HaveOccurred())
		Expect(e2e.WaitForDaemonSet(constants.OpenshiftNS, constants.LogfilesmetricexporterName)).To(Succeed())

		pods, err := e2e.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", constants.LabelK8sComponent, constants.LogfilesmetricexporterName),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pods.Items).ToNot(BeEmpty(), "Expected to find at least one LFME pod")

		podName := pods.Items[0].Name

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
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		metricsURL := fmt.Sprintf("https://localhost:%d/metrics", localPort)

		err = wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Second*30, true, func(context.Context) (done bool, err error) {
			resp, err := httpClient.Get(metricsURL)
			if err != nil {
				return false, nil
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return false, nil
			}
			out := string(body)
			log.V(5).Info("Polling metrics", "result", out)
			if !strings.Contains(out, "log_logged_bytes_total") {
				return false, nil
			}
			return regexp.MatchString(`log_logged_bytes_total{.*} [1-9][0-9]*`, out)
		})
		Expect(err).ToNot(HaveOccurred(), "Exp. to find log_logged_bytes_total being calculated")
	})
})
