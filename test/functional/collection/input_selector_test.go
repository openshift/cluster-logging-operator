package collection

import (
	"fmt"
	"runtime"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	"github.com/ViaQ/logerr/v2/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[Functional][Collection] InputSelector filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	logger := log.NewLogger("input-selector-testing")
	logger.Info("Running ", "filename", filename)

	const (
		otherFluentForward = "altFluentForward"
	)

	var (
		instance *functional.CollectorFunctionalFramework
	)
	//appNamespace2 := "application-ns2"
	appLabels1 := map[string]string{"name": "app1", "env": "env1"}
	appLabels2 := map[string]string{"name": "app1", "fallback": "env2"}

	AfterEach(func() {
		instance.Cleanup()
	})

	Describe("when CLF has input selectors to collect application logs", func() {
		Describe("from pods identified by labels", func() {

			It("should send logs from specific applications by using labels", func() {
				instance = functional.NewCollectorFunctionalFramework()
				instance.Labels = map[string]string{
					"name":     "app1",
					"env":      "env1",
					"fallback": "env2",
				}
				builder := functional.NewClusterLogForwarderBuilder(instance.Forwarder).
					FromInputWithVisitor("application-logs1",
						func(spec *logging.InputSpec) {
							spec.Application = &logging.Application{
								Selector: &metav1.LabelSelector{
									MatchLabels: appLabels1,
								},
							}
						},
					).Named("app-1").
					ToFluentForwardOutput()
				builder.FromInputWithVisitor("application-logs2",
					func(spec *logging.InputSpec) {
						spec.Application = &logging.Application{
							Selector: &metav1.LabelSelector{
								MatchLabels: appLabels2,
							},
						}
					},
				).Named("app-2").
					ToOutputWithVisitor(
						func(spec *logging.OutputSpec) {
							spec.Type = logging.OutputTypeFluentdForward
							spec.URL = "tcp://0.0.0.0:24225"
						}, otherFluentForward)

				Expect(instance.Deploy()).To(BeNil())
				Expect(instance.WritesApplicationLogs(1)).To(Succeed(), "Expected no errors writing log messages")

				logs, err := instance.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
				Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeFluentdForward, err)
				Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeFluentdForward)

				// verify only appLabels1 logs appear in Application logs
				for _, msg := range logs {
					logger.V(3).Info("Print", "msg", msg)
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("name", "app1"))
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("env", "env1"))
				}

				logs, err = instance.ReadApplicationLogsFrom(otherFluentForward)
				Expect(err).To(BeNil(), "Error fetching logs from %s: %v", otherFluentForward, err)
				Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", otherFluentForward)
				// verify only appLabels2 logs appear in Application logs
				for _, msg := range logs {
					logger.V(3).Info("Print", "msg", msg)
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("name", "app1"))
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("fallback", "env2"))
				}
			})
		})
		Describe("from pods identified by labels and namespaces", func() {
			It("should send logs with labels name:app1 and env:env1 from namespace application-ns1 to fluentd only", func() {

				instance = functional.NewCollectorFunctionalFramework()
				instance.Labels = map[string]string{
					"name": "app1",
					"env":  "env1",
				}
				functional.NewClusterLogForwarderBuilder(instance.Forwarder).
					FromInputWithVisitor("application-logs",
						func(spec *logging.InputSpec) {
							spec.Application = &logging.Application{
								Namespaces: []string{instance.Namespace},
								Selector: &metav1.LabelSelector{
									MatchLabels: appLabels1,
								},
							}
						},
					).
					ToFluentForwardOutput()

				Expect(instance.Deploy()).To(BeNil())
				Expect(instance.WritesApplicationLogs(1)).To(Succeed(), "Expected no errors writing log messages")

				msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "Here is my message", false)
				filepath := fmt.Sprintf("/var/log/pods/%s_%s_%s/%s/0.log", "myname", "application1", "12345", "thecontainer")
				Expect(instance.WriteMessagesToLog(msg, 1, filepath)).To(Succeed(), "Expected no errors writing log messages")

				logs, err := instance.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
				Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeFluentdForward, err)
				Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeFluentdForward)

				// verify only appLabels1 logs appear in Application logs
				for _, msg := range logs {
					logger.V(3).Info("Print", "msg", msg)
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("name", "app1"))
					Expect(msg.Kubernetes.Labels).Should(HaveKeyWithValue("env", "env1"))
					Expect(msg.Message).To(Not(ContainSubstring("Here is my message")), "Found an unexpected long entry")
				}
			})
		})

	})

})
