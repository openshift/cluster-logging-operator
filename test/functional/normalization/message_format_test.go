package normalization

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][LogForwarding][Normalization] tests for message format", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		framework.Labels = map[string]string{
			"app.kubernetes.io/name": "somevalue",
			"foo.bar":                "a123",
		}
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToElasticSearchOutput().
			FromInput(obs.InputTypeAudit).
			ToElasticSearchOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should parse application log format correctly", func() {

		// Log message data
		message := "Functional test message"
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

		// Template expected as output Log
		var outputLogTemplate = functional.NewApplicationLogTemplate()
		outputLogTemplate.Timestamp = nanoTime
		outputLogTemplate.Message = fmt.Sprintf("regex:^%s.*$", message)
		outputLogTemplate.Level = "*"

		// Write log line as input
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		// Read line from Log Forward output
		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
		labelRegex := "^([a-zA-Z0-9_]*)$"
		Expect(outputTestLog.Kubernetes.Labels).To(HaveKey(MatchRegexp(labelRegex)))
		Expect(outputTestLog.Kubernetes.NamespaceLabels).To(HaveKey(MatchRegexp(labelRegex)))
	})

	It("should parse application log format correctly if log message contains 'stdout','stderr' (Bug 1889595)", func() {

		// Log message data
		message := "Functional test stdout stderr message "
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

		// Template expected as output Log
		var outputLogTemplate = functional.NewApplicationLogTemplate()
		outputLogTemplate.Timestamp = nanoTime
		outputLogTemplate.Message = fmt.Sprintf("regex:^%s.*$", message)
		outputLogTemplate.Level = "*"

		// Write log line as stdout
		applicationLogLine := fmt.Sprintf("%s stdout F %s $n", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		// Write log line as stderr
		applicationLogLine = fmt.Sprintf("%s stderr F %s $n", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))

		outputLogTemplate.Kubernetes.ContainerStream = "stderr"
		outputTestLog = logs[1]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	})

})
