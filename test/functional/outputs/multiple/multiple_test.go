package multiple

import (
	"sort"

	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var _ = Describe("[Functional][Outputs][Multiple] tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("LOG-1575", func() {
		It("should fix sending JSON logs to syslog and elasticsearch without error", func() {
			pipelineBuilder := testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication)
			pipelineBuilder.ToElasticSearchOutput()
			pipelineBuilder.ToSyslogOutput(obs.SyslogRFC5424, func(output *obs.OutputSpec) {
				output.Syslog.Severity = "informational"
				output.Syslog.AppName = "myapp"
			})
			Expect(framework.Deploy()).To(BeNil())

			//seed ES logstore to have non-structured message entry
			appMsg := "docker:fake non-json message"
			crioMsg := functional.NewCRIOLogMessage("2013-03-28T14:36:03.243000+00:00", appMsg, false)
			Expect(framework.WriteMessagesToApplicationLog(crioMsg, 1)).To(BeNil())

			appMsg = `{"t":{"$date":"2021-08-22T11:30:24.378+00:00"},"s":"I",  "c":"NETWORK",  "id":22943,   "ctx":"listener","msg":"Connection accepted","attr":{"remote":"127.0.0.1:36046","connectionId":624876,"connectionCount":13}}`
			crioMsg = functional.NewCRIOLogMessage("2013-04-28T14:36:03.243000+00:00", appMsg, false)
			Expect(framework.WriteMessagesToApplicationLog(crioMsg, 1)).To(BeNil())

			// Read line from Syslog output
			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).To(HaveLen(2), "Expected syslog to have received all the records")

			raw, err := framework.GetLogsFromElasticSearch(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			Expect(logs).To(HaveLen(2), "Expected Elasticsearch to have received all the records")
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].TimestampLegacy.Before(logs[j].TimestampLegacy)
			})
			Expect(logs[1].Message).To(Equal(appMsg))
		})
	})

	Context("LOG-3640", func() {
		It("should send parsed JSON logs to different outputs when using multiple pipelines", func() {
			builder := testruntime.NewClusterLogForwarderBuilder(framework.Forwarder)
			builder.FromInput(obs.InputTypeApplication).WithParseJson().Named("one").
				ToElasticSearchOutput(func(output *obs.OutputSpec) {
					output.Elasticsearch.URL = "http://0.0.0.0:9200"
					output.Elasticsearch.Index = "foo"
				})

			builder.FromInput(obs.InputTypeApplication).WithParseJson().Named("two").
				ToOutputWithVisitor(func(output *obs.OutputSpec) {
					output.Type = obs.OutputTypeElasticsearch
					output.Elasticsearch = &obs.Elasticsearch{
						URLSpec: obs.URLSpec{
							URL: "http://0.0.0.0:9800",
						},
						Index:   "foo",
						Version: 8,
					}
				}, "other-es")

			builder.FromInput(obs.InputTypeApplication).WithParseJson().Named("three").
				ToHttpOutput()

			Expect(framework.Deploy()).To(BeNil())

			appMsg := `{"key":"value"}`
			crioMsg := functional.NewCRIOLogMessage("2013-04-28T14:36:03.243000+00:00", appMsg, false)
			Expect(framework.WriteMessagesToApplicationLog(crioMsg, 1)).To(BeNil())

			otherLogs, err := framework.GetLogsFromElasticSearchIndex("other-es", "foo",
				functional.Option{
					Name:  "port",
					Value: "9800",
				},
			)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(otherLogs).To(HaveLen(1), "Expected syslog to have received all the records")

			esLogs, err := framework.GetLogsFromElasticSearchIndex(string(obs.OutputTypeElasticsearch), "foo")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(esLogs).To(HaveLen(1))

			httpLogs, err := framework.ReadLogsFrom(string(obs.OutputTypeHTTP), string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(httpLogs).To(HaveLen(1))

			for output, raw := range map[string][]string{"other-es": otherLogs, "elasticsearch": esLogs, "http": httpLogs} {
				// Parse log line
				var logs []types.ApplicationLog
				err = types.StrictlyParseLogsFromSlice(raw, &logs)
				Expect(err).To(BeNil(), "Expected no errors parsing the logs for %s", output)
				Expect(logs[0].Structured).To(Not(BeEmpty()), "Expected structured logs for", output)
			}
		})
	})
})
