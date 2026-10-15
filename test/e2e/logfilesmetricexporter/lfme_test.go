package logfilesmetricexporter

import (
	"context"
	_ "embed"
	"fmt"
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
)

//go:embed valid.yaml
var validCR string

//go:embed invalid.yaml
var inValidCR string

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

		// Wait for at least one pod to be running and ready
		var podName string
		log.V(3).Info("Waiting for LFME pod to be ready")
		err = wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Second*60, true, func(ctx context.Context) (bool, error) {
			pods, err := e2e.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", constants.LabelK8sComponent, constants.LogfilesmetricexporterName),
			})
			if err != nil {
				log.V(3).Info("Failed to list pods", "error", err)
				return false, nil
			}
			if len(pods.Items) == 0 {
				log.V(3).Info("No LFME pods found yet")
				return false, nil
			}

			log.V(3).Info("Checking pod readiness", "podCount", len(pods.Items))
			// Find a running and ready pod
			for _, pod := range pods.Items {
				log.V(3).Info("Pod status", "podName", pod.Name, "phase", pod.Status.Phase)
				if pod.Status.Phase == "Running" {
					for _, condition := range pod.Status.Conditions {
						if condition.Type == "Ready" {
							log.V(3).Info("Pod Ready condition", "podName", pod.Name, "status", condition.Status, "reason", condition.Reason, "message", condition.Message)
							if condition.Status == "True" {
								podName = pod.Name
								log.V(3).Info("Found ready pod", "podName", podName)
								return true, nil
							}
						}
					}
				}
			}
			log.V(3).Info("Waiting for pod to be ready", "phase", pods.Items[0].Status.Phase, "podName", pods.Items[0].Name)
			return false, nil
		})
		Expect(err).ToNot(HaveOccurred(), "Expected to find at least one ready LFME pod")

		log.V(3).Info("Pod is ready, waiting 5 seconds for metrics collection to start", "podName", podName)
		// Give the exporter a moment to start collecting logs after pod is ready
		time.Sleep(5 * time.Second)
		log.V(3).Info("Starting metrics polling", "podName", podName)

		// Use Kubernetes API proxy to access pod's metrics endpoint
		// The metrics endpoint serves HTTPS, so we need to use https: prefix
		// Format: pods/https:{podName}:{port}/proxy/{path}
		req := e2e.KubeClient.CoreV1().RESTClient().Get().
			Namespace(constants.OpenshiftNS).
			Resource("pods").
			SubResource("proxy").
			Name(fmt.Sprintf("https:%s:2112", podName)).
			Suffix("metrics")

		var attempt int
		err = wait.PollUntilContextTimeout(context.TODO(), time.Second*2, time.Second*60, true, func(ctx context.Context) (done bool, err error) {
			attempt++
			log.V(3).Info("Polling metrics endpoint", "attempt", attempt, "podName", podName)

			// Execute the request through the Kubernetes proxy
			result := req.Do(ctx)
			body, err := result.Raw()
			if err != nil {
				log.V(3).Info("Failed to fetch metrics (will retry)", "error", err, "attempt", attempt, "podName", podName)
				return false, nil
			}

			out := string(body)
			log.V(3).Info("Received metrics response", "attempt", attempt, "bodyLength", len(out))

			if !strings.Contains(out, "log_logged_bytes_total") {
				log.V(3).Info("Metric log_logged_bytes_total not found yet (will retry)", "attempt", attempt, "bodyPreview", out[:min(200, len(out))])
				return false, nil
			}

			matched, err := regexp.MatchString(`log_logged_bytes_total{.*} [1-9][0-9]*`, out)
			if err != nil {
				log.V(3).Info("Regex match error", "error", err, "attempt", attempt)
				return false, nil
			}

			if matched {
				log.V(3).Info("Successfully found log_logged_bytes_total metric with non-zero value", "attempt", attempt)
				return true, nil
			}

			log.V(3).Info("Metric found but value is zero (will retry)", "attempt", attempt)
			return false, nil
		})
		Expect(err).ToNot(HaveOccurred(), "Exp. to find log_logged_bytes_total being calculated")
	})
})
