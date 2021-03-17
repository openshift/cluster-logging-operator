package normalization

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	"strings"
)

var _ = Describe("[Functional][Normalization][Json] Parse json format log", func() {

	var (
		framework *functional.FluentdFunctionalFramework
		// json message
		jsonMsg = "{\\\"name\\\":\\\"fred\\\",\\\"home\\\":\\\"bedrock\\\"}"
	)

	BeforeEach(func() {

		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when parsing json logs as un-structured", func() {
		It("should be logged as a message", func() {
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesJsonApplicationLogs(jsonMsg, 1)).To(BeNil())

			raw, err := framework.ReadApplicationLogsFrom("fluentForward")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			logs, err := types.ParseLogs(raw)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputLogTemplate := functional.NewLogTemplate()
			outputLogTemplate.Message = strings.Replace(jsonMsg, "\\", "", -1)
			outputTestLog := logs[0]
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})

	Context("when parsing json logs as structured", func() {
		It("should be logged as a message", func() {

			framework.SetFluentConfigFileName("fluentd_structured_configuration.txt")
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesJsonApplicationLogs(jsonMsg, 1)).To(BeNil())

			raw, err := framework.ReadApplicationLogsFrom("fluentForward")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Compare to expected template
			outputLogTemplate := functional.NewLogTemplate()
			outputLogTemplate.Message = strings.Replace(jsonMsg, "\\", "", -1)

			// Verify structured field
			var jsonRaw []map[string]interface{}
			err = json.Unmarshal([]byte(raw), &jsonRaw)
			Expect(err).To(BeNil(), "Expected no errors Unmarshal Log")
			jsonMessageMarshal,err  := json.Marshal(jsonRaw[0]["structured"])
			Expect(err).To(BeNil(), "Expected no errors Marshal Message")
			Expect(string(jsonMessageMarshal)).To(MatchJSON(outputLogTemplate.Message))

			// Verify all other fields
			logs, err := types.ParseLogs(raw)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := logs[0]
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})

	Context("when parsing regular logs as structured", func() {
		It("should be logged as a message", func() {

			framework.SetFluentConfigFileName("fluentd_structured_configuration.txt")
			Expect(framework.Deploy()).To(BeNil())
			Expect(framework.WritesApplicationLogs( 1)).To(BeNil())

			raw, err := framework.ReadApplicationLogsFrom("fluentForward")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Compare to expected template
			outputLogTemplate := functional.NewLogTemplate()

			// Verify structured field is empty
			var jsonRaw []map[string]interface{}
			err = json.Unmarshal([]byte(raw), &jsonRaw)
			Expect(err).To(BeNil(), "Expected no errors Unmarshal Log")
			jsonMessageMarshal,err  := json.Marshal(jsonRaw[0]["structured"])
			Expect(err).To(BeNil(), "Expected no errors Marshal Message")
			Expect(string(jsonMessageMarshal)).To(MatchJSON("{}"))

			// Verify all other fields
			logs, err := types.ParseLogs(raw)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := logs[0]
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})

})
