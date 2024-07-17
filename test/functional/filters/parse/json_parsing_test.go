package parse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	Json = `
   {
     "a": "Alpha",
     "b": true,
     "c": 12345,
     "e": {
       "one": 1,
       "two": 2,
       "three": [
         3
       ],
       "four": [
         0,
         1,
         2,
         3,
         4
       ]
     },
     "h": {
       "a": {
         "b": {
           "c": {
             "d": {
               "e": {
                 "f": {
                   "g": 1
                 }
               }
             }
           }
         }
       }
     }
   }
   `
)

var _ = Describe("[functional][filters][parse] Json log parsing", func() {
	var (
		framework       *functional.CollectorFunctionalFramework
		expected        map[string]interface{}
		empty           map[string]interface{}
		expectedMessage string
		normalizeJson   = func(json string) string {
			json = strings.TrimSpace(strings.ReplaceAll(json, "\n", ""))
			json = strings.ReplaceAll(json, "\t", "")
			return strings.ReplaceAll(json, " ", "")
		}
	)

	BeforeEach(func() {
		empty = map[string]interface{}{}
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			WithParseJson().
			ToHttpOutput()

		expectedMessage = normalizeJson(Json)
	})
	AfterEach(func() {
		framework.Cleanup()
	})
	It("should parse json message into structured", func() {
		ExpectOK(json.Unmarshal([]byte(Json), &expected))
		ExpectOK(framework.Deploy())

		// Log message data
		applicationLogLine := functional.CreateAppLogFromJson(Json)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		log.V(2).Info("Received", "Message", logs[0].Message)
		Expect(logs[0].Message).To(BeEmpty())

		same := cmp.Equal(logs[0].Structured, expected)
		if !same {
			diff := cmp.Diff(logs[0].Structured, expected)
			fmt.Printf("diff %s\n", diff)
		}
		Expect(same).To(BeTrue(), "parsed json message not matching")
	})

	It("should not parse non json message into structured", func() {
		Expect(framework.Deploy()).To(BeNil())

		// Log message data
		message := "Functional test message"
		timestamp := "2020-11-04T18:13:59.061892+00:00"

		applicationLogLine := fmt.Sprintf("%s stdout F %s", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		same := cmp.Equal(logs[0].Structured, empty)
		if !same {
			diff := cmp.Diff(logs[0].Structured, empty)
			log.V(0).Info("Parsed json not as expected", "diff", diff)
		}
		Expect(same).To(BeFalse(), "parsed json message not matching")
		Expect(logs[0].Message).To(Equal(message), "received message not matching")
	})
	It("should not parse json if not configured", func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput()
		ExpectOK(framework.Deploy())

		applicationLogLine := functional.CreateAppLogFromJson(expectedMessage)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		Expect(logs[0].Structured).To(BeNil(), "expected nil structured field")
		Expect(logs[0].Message).To(Equal(expectedMessage), "received message not matching")
	})
	It("should not parse invalid json message into structured", func() {
		Expect(framework.Deploy()).To(BeNil())

		// Log message data
		invalidJson := `{"key":"v}`
		timestamp := "2020-11-04T18:13:59.061892+00:00"

		expectedMessage := invalidJson
		applicationLogLine := fmt.Sprintf("%s stdout F %s", timestamp, expectedMessage)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		same := cmp.Equal(logs[0].Structured, empty)
		if !same {
			diff := cmp.Diff(logs[0].Structured, empty)
			log.V(3).Info("Parsed json not as expected", "diff", diff)
		}
		Expect(logs[0].Message).To(Equal(expectedMessage), "received message not matching")
	})

	It("should verify LOG-2105 parses json message into structured field and writes to Elasticsearch", func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			WithParseJson().
			ToElasticSearchOutput(func(spec *obs.OutputSpec) {
				spec.Elasticsearch.Index = "{.kubernetes.namespace_name}-write"
			})

		ExpectOK(framework.Deploy())

		// Log message data
		sample := `{"@timestamp":"2021-12-14T15:12:47.645Z","message":"Building mime message for recipient 'auser@somedomain.com' and sender 'Sympany <no-reply@somedomain>'.","level":"DEBUG","logger_name":"ch.sympany.backend.notificationservice.mail.MailServiceBean","thread_name":"default task-4"}`
		expectedMessage = normalizeJson(sample)
		expected = map[string]interface{}{}
		_ = json.Unmarshal([]byte(sample), &expected)

		applicationLogLine := functional.CreateAppLogFromJson(sample)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.GetLogsFromElasticSearchIndex(string(obs.OutputTypeElasticsearch), fmt.Sprintf("%s-write", framework.Namespace))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(raw).To(Not(BeEmpty()))

		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogsFromSlice(raw, &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs).To(HaveLen(1), "Expected to receive the log message")
		Expect(logs[0].Structured).To(Equal(expected), "structured field with parsed json message not matching")
		log.V(2).Info("Received", "Message", logs[0])
		Expect(logs[0].Message).To(BeEmpty())
	})
})
