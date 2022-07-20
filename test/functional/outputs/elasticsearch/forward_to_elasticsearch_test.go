package elasticsearch

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	"sort"
	"time"
)

var _ = Describe("[Functional][Outputs][ElasticSearch] FluentdForward Output to ElasticSearch", func() {

	var (
		framework *functional.CollectorFunctionalFramework

		// Template expected as output Log
		outputLogTemplate = functional.NewApplicationLogTemplate()
	)

	BeforeEach(func() {
		outputLogTemplate.ViaqIndexName = "app-write"
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to ElasticSearch "+functional.ElasticSearchTag+" protocol", func() {
		It("should  be compatible", func() {
			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
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
		It("should work well for valid utf-8 and replace not utf-8", func() {
			timestamp := functional.CRIOTime(time.Now())
			ukr := "привіт "
			jp := "こんにちは "
			ch := "你好"
			msg := functional.NewCRIOLogMessage(timestamp, ukr+jp+ch, false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			Expect(framework.WriteMessagesWithNotUTF8SymbolsToLog()).To(BeNil())

			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))
			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogs(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(len(logs)).To(Equal(2))
			//sort log by time before matching
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].Timestamp.Before(logs[j].Timestamp)
			})

			Expect(logs[0].Message).To(Equal(ukr + jp + ch))
			Expect(logs[1].Message).To(Equal("������������"))
		})
	})
})
