package normalization

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[LogForwarding] Functional tests for message format", func() {

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput().
			FromInput(logging.InputNameAudit).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should parse k8s audit log format correctly", func() {
		// Log message data
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

		// Define a template for test format (used for input, and expected output)
		var outputLogTemplate = types.K8sAuditLog{
			AuditLogCommon: types.AuditLogCommon{
				Kind:             "Event",
				Hostname:         functional.FunctionalNodeName,
				LogType:          "audit",
				ViaqIndexName:    "audit-write",
				Level:            "debug",
				Timestamp:        time.Time{},
				ViaqMsgID:        "*",
				PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
			},
			K8SAuditLevel: "debug",
		}
		// Template expected as output Log
		k8sAuditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"debug"}`, functional.CRIOTime(nanoTime))
		Expect(framework.WriteMessagesTok8sAuditLog(k8sAuditLogLine, 10)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.K8sAuditLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	})

	It("should parse linux audit log format correctly", func() {
		// Log message data
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		testTime, _ := time.Parse(time.RFC3339Nano, timestamp)
		auditLogLine := functional.NewAuditHostLog(testTime)
		// Template expected as output Log
		var outputLogTemplate = types.LinuxAuditLog{
			Message:       auditLogLine,
			LogType:       "audit",
			Hostname:      functional.FunctionalNodeName,
			ViaqIndexName: "audit-write",
			AuditLinux: types.AuditLinux{
				Type:     "DAEMON_START",
				RecordID: "*",
			},
			Timestamp:        testTime,
			ViaqMsgID:        "*",
			PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
		}
		// Write log line as input to fluentd
		Expect(framework.WriteMessagesToAuditLog(auditLogLine, 10)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.LinuxAuditLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
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

		// Write log line as input to fluentd
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	},
		Entry("should recognize a WARN message", "warn", "Warning: failed to query journal: Bad message OS Error 74"),
		Entry("should recognize an INFO message", "info", "I0920 14:22:00.089385       1 scheduler.go:592] \"Successfully bound pod to node\" pod=\"openshift-marketplace/community-operators-qrs99\" node=\"ip-10-0-215-216.us-east-2.compute.internal\" evaluatedNodes=6 feasibleNodes=3"),
		Entry("should recognize an ERROR message", "error", "E0427 02:47:01.619035 1 authentication.go:53] Unable to authenticate the request due to an error: invalid bearer token, context canceled"),
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

		// Write log line as input to fluentd
		applicationLogLine := fmt.Sprintf("%s stdout F %s $n", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
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

		// Write log line as input to fluentd
		applicationLogLine := fmt.Sprintf("%s stdout F %s $n", timestamp, message)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	})

	It("should parse ovn audit log correctly", func() {
		// Log message data
		level := "info"
		ovnLogLine := functional.NewOVNAuditLog(time.Now())

		// Template expected as output Log
		var outputLogTemplate = types.OVNAuditLog{
			Message:          ovnLogLine,
			Level:            level,
			Hostname:         functional.FunctionalNodeName,
			Timestamp:        time.Time{},
			LogType:          "audit",
			ViaqIndexName:    "audit-write",
			ViaqMsgID:        "*",
			PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
		}
		outputLogTemplate.PipelineMetadata.Collector.ReceivedAt = time.Time{}
		// Write log line as input to fluentd
		Expect(framework.WriteMessagesToOVNAuditLog(ovnLogLine, 10)).To(BeNil())
		// Read line from Log Forward output
		raw, err := framework.ReadOvnAuditLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		var logs []types.OVNAuditLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		ExpectOK(err)
		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
	})

})
