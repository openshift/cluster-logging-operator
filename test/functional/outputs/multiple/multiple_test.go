package multiple

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var _ = Describe("[Functional][Outputs][Multiple] tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	const SYSLOG_NAME = "asyslog"

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("LOG-1575", func() {
		It("should fix sending JSON logs to syslog and elasticsearch without error", func() {
			pipelineBuilder := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication)
			pipelineBuilder.ToElasticSearchOutput()
			pipelineBuilder.ToOutputWithVisitor(func(spec *logging.OutputSpec) {
				spec.Name = SYSLOG_NAME
				spec.Syslog = &logging.Syslog{
					Severity: "informational",
					AppName:  "myapp",
					RFC:      "RFC5424",
				}
			},
				logging.OutputTypeSyslog,
			)
			Expect(framework.Deploy()).To(BeNil())

			//seed ES logstore to have non-structured message entry
			appMsg := "docker:fake non-json message"
			crioMsg := functional.NewCRIOLogMessage("2013-03-28T14:36:03.243000+00:00", appMsg, false)
			Expect(framework.WriteMessagesToApplicationLog(crioMsg, 1)).To(BeNil())

			appMsg = `{"t":{"$date":"2021-08-22T11:30:24.378+00:00"},"s":"I",  "c":"NETWORK",  "id":22943,   "ctx":"listener","msg":"Connection accepted","attr":{"remote":"127.0.0.1:36046","connectionId":624876,"connectionCount":13}}`
			crioMsg = functional.NewCRIOLogMessage("2013-04-28T14:36:03.243000+00:00", appMsg, false)
			Expect(framework.WriteMessagesToApplicationLog(crioMsg, 1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(SYSLOG_NAME)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).To(HaveLen(2), "Expected syslog to have received all the records")

			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			Expect(logs).To(HaveLen(2), "Expected Elasticsearch to have received all the records")
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].Timestamp.Before(logs[j].Timestamp)
			})
			Expect(logs[1].Message).To(Equal(appMsg))
		})
	})
})
