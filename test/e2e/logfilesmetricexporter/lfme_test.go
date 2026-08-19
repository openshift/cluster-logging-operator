package logfilesmetricexporter

import (
	"context"
	_ "embed"
	"fmt"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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

		// Give the exporter a moment to start collecting logs
		time.Sleep(5 * time.Second)

		// Get a pod from the DaemonSet using the standard Kubernetes label
		pods, err := e2e.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", constants.LabelK8sComponent, constants.LogfilesmetricexporterName),
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pods.Items).ToNot(BeEmpty(), "Expected to find at least one LFME pod")
		podName := pods.Items[0].Name

		// Use Kubernetes API proxy to access pod's metrics endpoint
		// The metrics endpoint serves HTTPS, so we need to use https: prefix
		// Format: pods/https:{podName}:{port}/proxy/{path}
		req := e2e.KubeClient.CoreV1().RESTClient().Get().
			Namespace(constants.OpenshiftNS).
			Resource("pods").
			SubResource("proxy").
			Name(fmt.Sprintf("https:%s:2112", podName)).
			Suffix("metrics")

		err = wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Second*30, true, func(ctx context.Context) (done bool, err error) {
			// Execute the request through the Kubernetes proxy
			result := req.Do(ctx)
			body, err := result.Raw()
			if err != nil {
				log.V(5).Info("Failed to fetch metrics", "error", err)
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
