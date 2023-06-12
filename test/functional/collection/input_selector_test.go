package collection

import (
	"fmt"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"runtime"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("[Functional][Collection] InputSelector filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)

	const (
		otherOutput = "elasticsearch2"
	)

	var (
		instance *functional.CollectorFunctionalFramework
	)

	appLabels2 := map[string]string{"name": "app1", "fallback": "env2"}

	AfterEach(func() {
		instance.Cleanup()
	})

	Describe("when CLF has input selectors to collect application logs", func() {
		Describe("from pods identified by labels", func() {
			appLabels1 := map[string]string{"name": "app1", "env": "env1", "local.test/logtype": "user"}
			It("should send logs from specific applications by using labels", func() {
				instance = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
				instance.Labels = map[string]string{
					"name":               "app1",
					"env":                "env1",
					"fallback":           "env2",
					"local.test/logtype": "user",
				}
				builder := functional.NewClusterLogForwarderBuilder(instance.Forwarder).
					FromInputWithVisitor("application-logs1",
						func(spec *logging.InputSpec) {
							spec.Application = &logging.Application{
								Selector: &logging.LabelSelector{
									MatchLabels: appLabels1,
								},
							}
						},
					).Named("app-1").
					ToElasticSearchOutput()
				builder.FromInputWithVisitor("application-logs2",
					func(spec *logging.InputSpec) {
						spec.Application = &logging.Application{
							Selector: &logging.LabelSelector{
								MatchLabels: appLabels2,
							},
						}
					},
				).Named("app-2").
					ToOutputWithVisitor(
						func(spec *logging.OutputSpec) {
							spec.Name = otherOutput
							spec.Type = logging.OutputTypeElasticsearch
							spec.URL = "http://0.0.0.0:9201"
						}, otherOutput)

				Expect(instance.Deploy()).To(BeNil())
				Expect(instance.WritesApplicationLogs(1)).To(Succeed(), "Expected no errors writing log messages")

				logs, err := instance.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeElasticsearch, err)
				Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeElasticsearch)

				// verify only appLabels1 logs appear in Application logs
				for _, msg := range logs {
					log.V(3).Info("Print", "msg", msg)
					Expect(msg.Kubernetes.FlatLabels).Should(ContainElement("name=app1"))
					Expect(msg.Kubernetes.FlatLabels).Should(ContainElement("env=env1"))
				}

				logs, err = instance.ReadApplicationLogsFrom(otherOutput)
				Expect(err).To(BeNil(), "Error fetching logs from %s: %v", otherOutput, err)
				Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", otherOutput)
				// verify only appLabels2 logs appear in Application logs
				for _, msg := range logs {
					log.V(3).Info("Print", "msg", msg)
					Expect(msg.Kubernetes.FlatLabels).Should(ContainElement("name=app1"))
					Expect(msg.Kubernetes.FlatLabels).Should(ContainElement("fallback=env2"))
				}
			})
		})
		Describe("from pods identified by labels and namespaces", func() {
			It("should only send logs with labels name:app1 and env:env1 from namespace application-ns1", func() {
				appLabels1 := map[string]string{"name": "app1", "env": "env1"}
				instance = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
				instance.Labels = map[string]string{
					"name": "app1",
					"env":  "env1",
				}
				functional.NewClusterLogForwarderBuilder(instance.Forwarder).
					FromInputWithVisitor("application-logs",
						func(spec *logging.InputSpec) {
							spec.Application = &logging.Application{
								Namespaces: []string{instance.Namespace},
								Selector: &logging.LabelSelector{
									MatchLabels: appLabels1,
								},
							}
						},
					).
					ToElasticSearchOutput()

				Expect(instance.Deploy()).To(BeNil())
				Expect(instance.WritesApplicationLogs(1)).To(Succeed(), "Expected no errors writing log messages")

				msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "Here is my message", false)
				filepath := fmt.Sprintf("/var/log/pods/%s_%s_%s/%s/0.log", "myname", "application1", "12345", "thecontainer")
				Expect(instance.WriteMessagesToLog(msg, 1, filepath)).To(Succeed(), "Expected no errors writing log messages")

				logs, err := instance.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeElasticsearch, err)
				Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeElasticsearch)

				// verify only appLabels1 logs appear in Application logs
				for _, msg := range logs {
					log.V(3).Info("Print", "msg", msg)
					Expect(msg.Kubernetes.FlatLabels).Should(ContainElement("name=app1"))
					Expect(msg.Kubernetes.FlatLabels).Should(ContainElement("env=env1"))
					Expect(msg.Message).To(Not(ContainSubstring("Here is my message")), "Found an unexpected long entry")
				}
			})
		})

	})

})
