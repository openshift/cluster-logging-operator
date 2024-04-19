package normalization

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][LogForwarding][Normalization] tests for message format", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput().
			FromInput(logging.InputNameAudit).
			ToElasticSearchOutput()
		visitors := append(framework.AddOutputContainersVisitors(), func(builder *runtime.PodBuilder) error {
			builder.AddLabels(map[string]string{
				"app.kubernetes.io/name": "somevalue",
				"foo.bar":                "a123",
			})
			return nil
		})
		Expect(framework.DeployWithVisitors(visitors)).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("when evaluating an application message", func(expLevel, message string) {
		// Log message data
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

		// Template expected as output Log
		var outputLogTemplate = functional.NewApplicationLogTemplate()
		outputLogTemplate.Timestamp = nanoTime
		outputLogTemplate.Message = message
		outputLogTemplate.Level = expLevel

		applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	},
		Entry("should recognize a WARN message", "warn", "Warning: failed to query journal: Bad message OS Error 74"),
		Entry("should recognize an INFO message", "info", "I0920 14:22:00.089385       1 scheduler.go:592] \"Successfully bound pod to node\" pod=\"openshift-marketplace/community-operators-qrs99\" node=\"ip-10-0-215-216.us-east-2.compute.internal\" evaluatedNodes=6 feasibleNodes=3"),
		Entry("should recognize an ERROR message", "error", "E0427 02:47:01.619035 1 authentication.go:53] Unable to authenticate the request due to an error: invalid bearer token, context canceled"),
		Entry("should recognize an CRITICAL message", "critical", "CRITICAL:  Unable to connect to server"),
		Entry("should recognize an DEBUG message", "debug", "level=debug found the light"),
	)
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
		logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
		Expect(outputTestLog.Kubernetes.Labels).To(HaveKey(MatchRegexp("^([a-zA-Z0-9_]*)$")))
		Expect(outputTestLog.Kubernetes.Labels).ToNot(HaveKey(MatchRegexp("foo")))
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

		// Write log line as input
		applicationLogLine := fmt.Sprintf("%s stdout F %s $n", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	})

})
