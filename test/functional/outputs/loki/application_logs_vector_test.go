//go:build vector

package loki

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
)

var _ = Describe("[Functional][Outputs][Loki] Forwarding to Loki", func() {

	var (
		f *functional.CollectorFunctionalFramework
		l *loki.Receiver
	)

	BeforeEach(func() {
		f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		// Start a Loki server
		l = loki.NewReceiver(f.Namespace, "loki-server")
		Expect(l.Create(f.Test.Client)).To(Succeed())

		// Set up the common template forwarder configuration.
		f.Forwarder.Spec.Outputs = append(f.Forwarder.Spec.Outputs,
			logging.OutputSpec{
				Name: logging.OutputTypeLoki,
				Type: logging.OutputTypeLoki,
				URL:  l.InternalURL("").String(),
				OutputTypeSpec: logging.OutputTypeSpec{
					Loki: &logging.Loki{},
				},
			})
		f.Forwarder.Spec.Pipelines = append(f.Forwarder.Spec.Pipelines,
			logging.PipelineSpec{
				Name:       "functional-loki-pipeline_0_",
				OutputRefs: []string{logging.OutputTypeLoki},
				InputRefs:  []string{logging.InputNameApplication},
				Labels:     map[string]string{"logging": "logging-value"},
			})

	})

	AfterEach(func() {
		f.Cleanup()
	})

	Context("with vector not ordered events", func() {
		BeforeEach(func() {
			Expect(f.Deploy()).To(BeNil())
		})
		It("should accept not ordered event", func() {
			now := time.Now()
			tsNow := functional.CRIOTime(now)
			duration, _ := time.ParseDuration("-5.5h") //time back
			then := now.Add(duration)
			tsThen := then.UTC().Format(time.RFC3339Nano)
			msg := functional.NewFullCRIOLogMessage(tsNow, "Present days")
			msgOld := functional.NewFullCRIOLogMessage(tsThen, "A long time ago in a galaxy far, far away....")
			msgNew := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "Present days")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(Succeed())
			Expect(f.WriteMessagesToApplicationLog(msgOld, 1)).To(Succeed())
			Expect(f.WriteMessagesToApplicationLog(msgNew, 1)).To(Succeed())

			query := fmt.Sprintf(`{kubernetes_namespace_name=%q, kubernetes_pod_name=%q}`, f.Namespace, f.Name)
			result, err := l.QueryUntil(query, "", 3)
			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(len(result)).To(Equal(1))
			lines := result[0].Lines()
			Expect(len(lines)).To(Equal(3))
			Expect(strings.Contains(lines[0], "Present days")).To(BeTrue())
			Expect(strings.Contains(lines[1], "A long time ago in a galaxy far, far away....")).To(BeTrue())
			Expect(strings.Contains(lines[2], "Present days")).To(BeTrue())
		})
	})
	Context("labelKeys", func() {
		const myValue = "foobarvalue"
		It("should handle the configuration so the collector starts when label keys are defined that include slashes and dashes. Ref(LOG-4095, LOG-4460)", func() {
			f.Labels["app.kubernetes.io/name"] = myValue
			f.Labels["prefix-cloud_com_platform-stage"] = "dev"
			f.Forwarder.Spec.Outputs[0].Loki.LabelKeys = []string{
				"kubernetes.namespace_name",
				"kubernetes.pod_name",
				"kubernetes.labels.app.kubernetes.io/name",
				"kubernetes.labels.prefix-cloud_com_platform-stage",
			}
			Expect(f.Deploy()).To(BeNil())
			now := time.Now()
			tsNow := functional.CRIOTime(now)
			msg := functional.NewFullCRIOLogMessage(tsNow, "Present days")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(Succeed())

			query := fmt.Sprintf(`{kubernetes_labels_app_kubernetes_io_name=%q}`, myValue)
			result, err := l.QueryUntil(query, "", 1)
			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(len(result)).To(Equal(1))
			lines := result[0].Lines()
			Expect(len(lines)).To(Equal(1))

			want := map[string]string{
				"k8s_namespace_name":                                f.Namespace,
				"k8s_pod_name":                                      f.Pod.Name,
				"k8s_node_name":                                     f.Pod.Spec.NodeName,
				"kubernetes_namespace_name":                         f.Namespace,
				"kubernetes_pod_name":                               f.Pod.Name,
				"kubernetes_labels_app_kubernetes_io_name":          myValue,
				"kubernetes_labels_prefix_cloud_com_platform_stage": "dev",
				"kubernetes_host":                                   f.Pod.Spec.NodeName,
			}
			labels := result[0].Stream
			Expect(len(labels)).To(Equal(8))
			Expect(labels).To(BeEquivalentTo(want))
		})

		It("should add all otel equivalent default labels when loki.LabelKeys are not defined", func() {
			Expect(f.Deploy()).To(BeNil())
			now := time.Now()
			tsNow := functional.CRIOTime(now)
			msg := functional.NewFullCRIOLogMessage(tsNow, "Present days")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(Succeed())

			query := fmt.Sprintf(`{openshift_log_type=%q}`, logging.InputNameApplication)
			result, err := l.QueryUntil(query, "", 1)
			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(len(result)).To(Equal(1))
			lines := result[0].Lines()
			Expect(len(lines)).To(Equal(1))

			want := map[string]string{
				"k8s_container_name":        f.Pod.Spec.Containers[0].Name,
				"k8s_namespace_name":        f.Namespace,
				"k8s_pod_name":              f.Pod.Name,
				"k8s_node_name":             f.Pod.Spec.NodeName,
				"kubernetes_container_name": f.Pod.Spec.Containers[0].Name,
				"kubernetes_namespace_name": f.Namespace,
				"kubernetes_pod_name":       f.Pod.Name,
				"kubernetes_host":           f.Pod.Spec.NodeName,
				"log_type":                  logging.InputNameApplication,
				"openshift_log_type":        logging.InputNameApplication,
			}
			labels := result[0].Stream
			Expect(len(labels)).To(Equal(10))
			Expect(labels).To(BeEquivalentTo(want))
		})
	})

})
