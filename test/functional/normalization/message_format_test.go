package normalization

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
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
		kind := fmt.Sprintf("audit(%.3f:24287)", float64(nanoTime.UnixNano())/float64(time.Second))

		// Define a template for test format (used for input, and expected output)
		var outputLogTemplate = types.K8sAuditLog{
			AuditLogCommon: types.AuditLogCommon{
				Kind:             kind,
				ViaqIndexName:    "audit-write",
				Level:            "info",
				Timestamp:        time.Time{},
				ViaqMsgID:        "*",
				PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
			},
		}
		// Template expected as output Log
		k8sAuditLogLine := "{\\\"kind\\\":\\\"" + kind + "\\\"}"
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
		msgType := "CWD"
		timestamp := "2013-03-28T14:36:03.243000+00:00"
		nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
		msg := fmt.Sprintf("audit(%.3f:24287)", float64(nanoTime.UnixNano())/float64(time.Second))
		auditLogLine := fmt.Sprintf("type=%s msg=%s:", msgType, msg)
		// Template expected as output Log
		var outputLogTemplate = types.LinuxAuditLog{
			Message:       auditLogLine,
			ViaqIndexName: "audit-write",
			AuditLinux: types.AuditLinux{
				Type:     msgType,
				RecordID: "*",
			},
			Timestamp:        nanoTime,
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

})
