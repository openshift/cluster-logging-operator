package loki

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
)

var _ = Describe("[Functional][Outputs][Loki] Forwarding application logs to Loki", func() {
	var (
		f      *functional.CollectorFunctionalFramework
		l      *loki.Receiver
		tsTime = time.Now()
		ts     = functional.CRIOTime(tsTime)

		containerTag = func(f *functional.CollectorFunctionalFramework) string {
			for _, s := range f.Pod.Status.ContainerStatuses {
				if s.Name == constants.CollectorName {
					return fmt.Sprintf("kubernetes.var.log.pods.%s_%s_%s.%s.0.log", f.Pod.Namespace, f.Pod.Name, f.Pod.UID, f.Pod.Spec.Containers[0].Name)
				}
			}
			Fail("Unable to find the container id to create a tag for the test")
			return ""
		}
	)

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFramework()
		l = loki.NewReceiver(f.Namespace, "loki-server").EnableMultiTenant()
		ExpectOK(l.Create(f.Test.Client))

		// Set up the common template forwarder configuration.
		f.Forwarder.Spec.Outputs = append(f.Forwarder.Spec.Outputs,
			logging.OutputSpec{
				Name: "loki",
				Type: "loki",
				URL:  l.InternalURL("").String(),
			})
		f.Forwarder.Spec.Pipelines = append(f.Forwarder.Spec.Pipelines,
			logging.PipelineSpec{
				OutputRefs: []string{"loki"},
				InputRefs:  []string{"application"},
				Labels:     map[string]string{"logging": "logging-value"},
			})
	})

	AfterEach(func() {
		f.Cleanup()
	})

	deploy := func(lokiSpec *logging.Loki) {
		f.Forwarder.Spec.Outputs[0].OutputTypeSpec = logging.OutputTypeSpec{Loki: lokiSpec}
		ExpectOK(f.DeployWithVisitor(func(p *runtime.PodBuilder) error {
			p.AddLabels(map[string]string{"k8s": "k8s-value"})
			return nil
		}))
	}

	It("forwards application logs with default labels and tenant", func() {
		deploy(nil)
		msg := functional.NewFullCRIOLogMessage(ts, "application log message")
		ExpectOK(f.WriteMessagesToApplicationLog(msg, 3))
		r, err := l.QueryUntil(`{log_type="application"}`, "application", 3)
		ExpectOK(err)

		// Check expected Loki labels
		Expect(r).NotTo(BeEmpty())
		labels := r[0].Stream
		delete(labels, "fluentd_thread") // Added by loki plugin.
		want := map[string]string{
			"log_type":                  "application",
			"kubernetes_host":           functional.FunctionalNodeName,
			"kubernetes_namespace_name": f.Namespace,
			"kubernetes_pod_name":       f.Name,
			"kubernetes_container_name": f.Pod.Spec.Containers[0].Name,
			"tag":                       containerTag(f),
		}
		Expect(labels).To(Equal(want))

		// Check expected log records
		records := r[0].Records()
		Expect(records).To(HaveLen(3))
		for _, record := range records {
			Expect(record["log_type"]).To(Equal("application"))
			Expect(record["message"]).To(Equal("application log message"))
			k := record["kubernetes"].(map[string]interface{})
			Expect(k["namespace_name"]).To(Equal(f.Namespace))
			Expect(k["pod_name"]).To(Equal(f.Name))
			// Timestamp will not match exactly, some sub-second digits are truncated.
			recordTime, err := time.Parse(time.RFC3339Nano, record["@timestamp"].(string))
			ExpectOK(err)
			diff := recordTime.Sub(tsTime)
			if diff < 0 {
				diff = -diff
			}
			Expect(diff).To(BeNumerically("<", time.Millisecond))
			k8sLabels := k["labels"].(map[string]interface{})
			Expect(k8sLabels["k8s"]).To(Equal("k8s-value"))
		}
	})

	It("forwards application logs with custom Loki labels", func() {
		deploy(&logging.Loki{LabelKeys: []string{
			"kubernetes.labels.k8s",
			"openshift.labels.logging",
			"kubernetes.container_name",
		}})
		msg := functional.NewFullCRIOLogMessage(ts, "application log message")
		ExpectOK(f.WriteMessagesToApplicationLog(msg, 1))

		// Verify we can query by Loki labels
		query := fmt.Sprintf(`{kubernetes_labels_k8s=%q, openshift_labels_logging=%q}`, "k8s-value", "logging-value")
		r, err := l.QueryUntil(query, "application", 1)
		ExpectOK(err, query)
		Expect(r).NotTo(BeEmpty())
		records := r[0].Records()
		Expect(records).To(HaveLen(1))
		Expect(records[0]["message"]).To(Equal("application log message"))

		want := map[string]string{
			"kubernetes_container_name": f.Pod.Spec.Containers[0].Name,
			"kubernetes_labels_k8s":     "k8s-value",
			"openshift_labels_logging":  "logging-value",
			"kubernetes_host":           functional.FunctionalNodeName,
			"tag":                       containerTag(f),
		}
		labels := r[0].Stream
		delete(labels, "fluentd_thread") // Added by loki plugin.
		Expect(labels).To(Equal(want))
	})
})
