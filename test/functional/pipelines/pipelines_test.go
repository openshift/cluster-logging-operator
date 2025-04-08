package pipelines

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/utils/set"
)

var _ = Describe("[Functional][Pipelines] when there are multiple pipelines", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("should send logs to the forward.Output logstore", func(inputTypes ...obs.InputType) {
		var option client.TestOption
		sources := set.New(inputTypes...)
		if sources.Has(obs.InputTypeInfrastructure) {
			option = client.UseInfraNamespaceTestOption
		}
		framework = functional.NewCollectorFunctionalFramework(option)

		writers := map[obs.InputType]func(int) error{
			obs.InputTypeAudit:          framework.WriteK8sAuditLog,
			obs.InputTypeInfrastructure: framework.WritesInfraContainerLogs,
			obs.InputTypeApplication:    framework.WritesApplicationLogs,
		}
		readers := map[obs.InputType]func(string) ([]string, error){
			obs.InputTypeAudit:          framework.ReadAuditLogsFrom,
			obs.InputTypeInfrastructure: framework.ReadInfrastructureLogsFrom,
			obs.InputTypeApplication:    framework.ReadRawApplicationLogsFrom,
		}

		builder := testruntime.NewClusterLogForwarderBuilder(framework.Forwarder)
		for _, source := range sources.SortedList() {
			builder.FromInput(source).Named("test-" + string(source)).ToElasticSearchOutput()
		}

		Expect(framework.Deploy()).To(BeNil())

		for _, source := range sources.SortedList() {
			log.V(1).Info("Writing log", "source", source)
			Expect(writers[source](1)).To(Succeed(), "Exp. to be able to write %s logs", source)

			logs, err := readers[source](string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Exp. no errors reading %s logs", source)
			Expect(logs).To(HaveLen(1), "Exp. to find logs")
			Expect(logs[0]).To(MatchRegexp(`log_type\"\:.*`+string(source)), "Exp. to find a log of type %s", source)
		}

	},
		Entry("when configured with audit and application sources", obs.InputTypeAudit, obs.InputTypeApplication),
		Entry("when configured with audit and infrastructure sources", obs.InputTypeAudit, obs.InputTypeInfrastructure),
	)
})
