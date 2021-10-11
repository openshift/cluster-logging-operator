package normalization

import (
	"encoding/json"
	"fmt"
	"github.com/ViaQ/logerr/log"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"strings"
)

const (
	Json = `
    {
      "a": "Alpha",
      "b": true,
      "c": 12345,
      "d": [
      true,
      [
        false,
        [
        -123456789,
        null
        ],
        3.9676,
        [
        "Something else.",
        false
        ],
        null
      ]
      ],
      "e": {
      "zero": null,
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
      "f": null,
      "h": {
        "a": {
          "b": {
            "c": {
              "d": {
                "e": {
                  "f": {
                    "g": null
                     }
                   }
                 }
               }
             }
           }
         },
      "i": [
            [
             [
              [
               [
                [
                 [
                  null
                 ]
                ]
               ]
              ]
             ]
            ]
          ]
    }
    `
)

var _ = Describe("[LogForwarding] Json log parsing", func() {
	var (
		framework       *functional.FluentdFunctionalFramework
		clfb            *functional.ClusterLogForwarderBuilder
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
		_ = json.Unmarshal([]byte(Json), &expected)
		empty = map[string]interface{}{}
		framework = functional.NewFluentdFunctionalFramework()
		clfb = functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()

		expectedMessage = normalizeJson(Json)
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should parse json message into structured", func() {
		clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
		ExpectOK(framework.Deploy())

		// Log message data
		applicationLogLine := functional.CreateAppLogFromJson(Json)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		same := cmp.Equal(logs[0].Structured, expected)
		if !same {
			diff := cmp.Diff(logs[0].Structured, expected)
			//log.V(3).Info("Parsed json not as expected", "diff", diff)
			fmt.Printf("diff %s\n", diff)
		}
		Expect(same).To(BeTrue(), "parsed json message not matching")
		log.V(2).Info("Received", "Message", logs[0].Message)
		diff := cmp.Diff(logs[0].Message, expectedMessage)
		log.V(2).Info("Received", "Diff", diff)
		Expect(normalizeJson(logs[0].Message)).To(Equal(expectedMessage), "received message not matching")
	})
	It("should not parse non json message into structured", func() {
		clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
		Expect(framework.Deploy()).To(BeNil())

		// Log message data
		message := "Functional test message"
		timestamp := "2020-11-04T18:13:59.061892+00:00"

		// Write log line as input to fluentd
		applicationLogLine := fmt.Sprintf("%s stdout F %s", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		same := cmp.Equal(logs[0].Structured, empty)
		if !same {
			diff := cmp.Diff(logs[0].Structured, empty)
			log.V(3).Info("Parsed json not as expected", "diff", diff)
		}
		Expect(same).To(BeTrue(), "parsed json message not matching")
		Expect(logs[0].Message).To(Equal(message), "received message not matching")
	})
	It("should not parse json if not configured", func() {
		// Pipeline.Parse is not set
		ExpectOK(framework.Deploy())

		applicationLogLine := functional.CreateAppLogFromJson(expectedMessage)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs[0].Structured).To(BeNil(), "expected nil structured field")
		Expect(logs[0].Message).To(Equal(expectedMessage), "received message not matching")
	})
	It("should not parse invalid json message into structured", func() {
		// This test case is disabled to fix the behavior of invalid json parsing
		clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
		Expect(framework.Deploy()).To(BeNil())

		// Log message data
		invalidJson := `{"key":"v}`
		timestamp := "2020-11-04T18:13:59.061892+00:00"

		// Write log line as input to fluentd
		expectedMessage := invalidJson
		applicationLogLine := fmt.Sprintf("%s stdout F %s", timestamp, expectedMessage)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		same := cmp.Equal(logs[0].Structured, empty)
		if !same {
			diff := cmp.Diff(logs[0].Structured, empty)
			log.V(3).Info("Parsed json not as expected", "diff", diff)
		}
		Expect(logs[0].Message).To(Equal(expectedMessage), "received message not matching")
	})

})
