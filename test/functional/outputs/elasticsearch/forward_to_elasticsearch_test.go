package elasticsearch

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][Outputs][ElasticSearch] Output to ElasticSearch", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)
	AfterEach(func() {
		framework.Cleanup()
	})
	DescribeTable("Functional test", func(collectorType logging.LogCollectionType) {
		// Template expected as output Log
		outputLogTemplate := functional.NewApplicationLogTemplate(collectorType)

		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(collectorType)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput()
		Expect(framework.Deploy()).To(BeNil())
		Expect(framework.WritesApplicationLogs(1)).To(BeNil())
		Context("when sending to ElasticSearch "+functional.ElasticSearchTag+" protocol", func() {
			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogs(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputTestLog := logs[0]
			outputLogTemplate.ViaqIndexName = ""
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	},
		Entry("with fluentd collector", logging.LogCollectionTypeFluentd),
		Entry("with vector collector", logging.LogCollectionTypeVector),
	)
})
