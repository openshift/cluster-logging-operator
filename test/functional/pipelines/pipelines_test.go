package pipelines

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"k8s.io/utils/strings/slices"
)

var _ = Describe("[Functional][Pipelines] when there are multiple pipelines", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("should send logs to the forward.Output logstore", func(sources ...string) {
		var option client.TestOption
		if slices.Contains(sources, logging.InputNameInfrastructure) {
			option = client.UseInfraNamespaceTestOption
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("Skipping test with vector because functional framework does not mock journal")
			}
		}
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType, option)

		writers := map[string]func(int) error{
			logging.InputNameAudit:          framework.WriteK8sAuditLog,
			logging.InputNameInfrastructure: framework.WritesInfraContainerLogs,
			logging.InputNameApplication:    framework.WritesApplicationLogs,
		}
		readers := map[string]func(string) ([]string, error){
			logging.InputNameAudit:          framework.ReadAuditLogsFrom,
			logging.InputNameInfrastructure: framework.ReadInfrastructureLogsFrom,
			logging.InputNameApplication:    framework.ReadRawApplicationLogsFrom,
		}

		builder := functional.NewClusterLogForwarderBuilder(framework.Forwarder)
		for _, source := range sources {
			builder.FromInput(source).Named("test-" + source).ToElasticSearchOutput()
		}

		Expect(framework.Deploy()).To(BeNil())

		for _, source := range sources {
			log.V(1).Info("Writing log", "source", source)
			Expect(writers[source](1)).To(Succeed(), "Exp. to be able to write %s logs", source)

			logs, err := readers[source](logging.OutputTypeElasticsearch)
			Expect(err).To(BeNil(), "Exp. no errors reading %s logs", source)
			Expect(logs).To(HaveLen(1), "Exp. to find logs")
			Expect(logs[0]).To(MatchRegexp(`log_type\"\:.*`+source), "Exp. to find a log of type %s", source)
		}

	},
		Entry("when configured with audit and application sources", logging.InputNameAudit, logging.InputNameApplication),
		Entry("when configured with audit and infrastructure sources", logging.InputNameAudit, logging.InputNameInfrastructure),
	)
})
