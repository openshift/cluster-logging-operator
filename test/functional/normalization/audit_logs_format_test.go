//go:build fluentd || vector
// +build fluentd vector

package normalization

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][LogForwarding][Normalization] message format tests for audit logs", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	Context("with fluentd", func() {
		if testfw.LogCollectionType == logging.LogCollectionTypeFluentd {
			BeforeEach(func() {
				framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeFluentd)
				functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameAudit).
					ToFluentForwardOutput()
				Expect(framework.Deploy()).To(BeNil())
			})

			It("should parse k8s audit log format correctly", func() {
				// Log message data
				timestamp := "2022-08-17T20:27:20.570375Z"
				nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

				// Define a template for test format (used for input, and expected output)
				var outputLogTemplate = types.K8sAuditLog{
					AuditLogCommon: types.AuditLogCommon{
						Kind:                     "Event",
						Hostname:                 functional.FunctionalNodeName,
						LogType:                  "audit",
						Level:                    "debug",
						Timestamp:                nanoTime,
						RequestReceivedTimestamp: nanoTime,
						ViaqMsgID:                "*",
						PipelineMetadata:         functional.TemplateForAnyPipelineMetadata,
					},
					K8SAuditLevel: "debug",
				}
				// Template expected as output Log
				k8sAuditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"debug"}`, timestamp)
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
					Message:  auditLogLine,
					LogType:  "audit",
					Hostname: functional.FunctionalNodeName,
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

			It("should parse ovn audit log correctly", func() {
				// Log message data
				level := "info"
				ovnLogLine := functional.NewOVNAuditLog(time.Now())

				// Template expected as output Log
				var outputLogTemplate = types.OVNAuditLog{
					Message:   ovnLogLine,
					Level:     level,
					Hostname:  functional.FunctionalNodeName,
					Timestamp: time.Time{},
					LogType:   "audit",
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt(""),
					},
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

			AfterEach(func() {
				framework.Cleanup()
			})
		}
	})

	Context("with vector", func() {
		if testfw.LogCollectionType == logging.LogCollectionTypeVector {
			BeforeEach(func() {
				framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
				functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameAudit).
					ToElasticSearchOutput()
				Expect(framework.Deploy()).To(BeNil())
			})

			//TODO: fix me when audit formatting is enabled
			It("should parse k8s audit log format correctly", func() {
				// Log message data
				timestamp := "2013-03-28T14:36:03.243000+00:00"
				nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

				// Define a template for test format (used for input, and expected output)
				//var outputLogTemplate = types.K8sAuditLog{
				//AuditLogCommon: types.AuditLogCommon{
				//	Kind:             "Event",
				//	Hostname:         functional.FunctionalNodeName,
				//	LogType:          "audit",
				//	Level:            "debug",
				//	Timestamp:        time.Time{},
				//	ViaqMsgID:        "*",
				//	PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
				//},
				//K8SAuditLevel: "debug",
				//}
				// Template expected as output Log
				k8sAuditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata"}`, functional.CRIOTime(nanoTime))
				Expect(framework.WriteMessagesTok8sAuditLog(k8sAuditLogLine, 10)).To(BeNil())
				// Read line from Log Forward output
				raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				//var logs []types.K8sAuditLog
				//err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
				//Expect(err).To(BeNil(), "Expected no errors parsing the logs: %v", raw)
				//// Compare to expected template
				//outputTestLog := logs[0]
				//Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
				results := strings.Join(raw, " ")
				Expect(results).To(MatchRegexp("kind.*Event.*level.*default.*k8s_audit_level.*Metadata"), "Message should contain the audit log: %v", raw)

			})
			//TODO: fix me when audit formatting is enabled
			It("should parse openshift audit log format correctly", func() {
				// Log message data
				timestamp := "2013-03-28T14:36:03.243000+00:00"
				nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

				// Define a template for test format (used for input, and expected output)
				//var outputLogTemplate = types.K8sAuditLog{
				//AuditLogCommon: types.AuditLogCommon{
				//	Kind:             "Event",
				//	Hostname:         functional.FunctionalNodeName,
				//	LogType:          "audit",
				//	Level:            "debug",
				//	Timestamp:        time.Time{},
				//	ViaqMsgID:        "*",
				//	PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
				//},
				//K8SAuditLevel: "debug",
				//}
				// Template expected as output Log
				auditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata"}`, functional.CRIOTime(nanoTime))
				Expect(framework.WriteMessagesToOpenshiftAuditLog(auditLogLine, 10)).To(BeNil())
				// Read line from Log Forward output
				raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				//var logs []types.K8sAuditLog
				//err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
				//Expect(err).To(BeNil(), "Expected no errors parsing the logs: %v", raw)
				//// Compare to expected template
				//outputTestLog := logs[0]
				//Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
				results := strings.Join(raw, " ")
				Expect(results).To(MatchRegexp("kind.*Event.*level.*default.*openshift_audit_level.*Metadata"), "Message should contain the audit log: %v", raw)

			})
			It("should parse oauth audit log format correctly", func() {
				// Log message data
				timestamp := "2013-03-28T14:36:03.243000+00:00"
				nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

				// Define a template for test format (used for input, and expected output)
				//var outputLogTemplate = types.K8sAuditLog{
				//AuditLogCommon: types.AuditLogCommon{
				//	Kind:             "Event",
				//	Hostname:         functional.FunctionalNodeName,
				//	LogType:          "audit",
				//	Level:            "debug",
				//	Timestamp:        time.Time{},
				//	ViaqMsgID:        "*",
				//	PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
				//},
				//K8SAuditLevel: "debug",
				//}
				// Template expected as output Log
				auditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"debug"}`, functional.CRIOTime(nanoTime))
				Expect(framework.WriteMessagesToOAuthAuditLog(auditLogLine, 10)).To(BeNil())
				// Read line from Log Forward output
				raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				//var logs []types.K8sAuditLog
				//err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
				//Expect(err).To(BeNil(), "Expected no errors parsing the logs: %v", raw)
				//// Compare to expected template
				//outputTestLog := logs[0]
				//Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
				results := strings.Join(raw, " ")
				Expect(results).To(MatchRegexp("kind.*Event.*level.*debug"), "Message should contain the audit log: %v", raw)

			})
			//TODO: fix me when audit formatting is enabled
			It("should parse linux audit log format correctly", func() {
				// Log message data
				timestamp := "2013-03-28T14:36:03.243000+00:00"
				testTime, _ := time.Parse(time.RFC3339Nano, timestamp)
				auditLogLine := functional.NewAuditHostLog(testTime)
				//// Template expected as output Log
				//var outputLogTemplate = types.LinuxAuditLog{
				//	Message:  auditLogLine,
				//	LogType:  "audit",
				//	Hostname: functional.FunctionalNodeName,
				//	AuditLinux: types.AuditLinux{
				//		Type:     "DAEMON_START",
				//		RecordID: "*",
				//	},
				//	Timestamp:        testTime,
				//	ViaqMsgID:        "*",
				//	PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
				//}
				// Write log line as input to fluentd
				Expect(framework.WriteMessagesToAuditLog(auditLogLine, 10)).To(BeNil())
				// Read line from Log Forward output
				raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				//var logs []types.LinuxAuditLog
				//err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
				//Expect(err).To(BeNil(), "Expected no errors parsing the logs")
				//// Compare to expected template
				//outputTestLog := logs[0]
				//Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
				results := strings.Join(raw, " ")
				Expect(results).To(MatchRegexp("format=enriched kernel="), "Message should contain the audit log: %v", raw)
			})
			//TODO: fix me when audit formatting is enabled
			It("should parse ovn audit log correctly", func() {
				// Log message data
				level := "info"
				ovnLogLine := functional.NewOVNAuditLog(time.Now())

				// Template expected as output Log
				var outputLogTemplate = types.OVNAuditLog{
					Message:   ovnLogLine,
					Level:     level,
					Hostname:  functional.FunctionalNodeName,
					Timestamp: time.Time{},
					LogType:   "audit",
					Openshift: types.OpenshiftMeta{
						Sequence: types.NewOptionalInt(""),
					},
					ViaqMsgID:        "*",
					PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
				}
				outputLogTemplate.PipelineMetadata.Collector.ReceivedAt = time.Time{}
				// Write log line as input to fluentd
				Expect(framework.WriteMessagesToOVNAuditLog(ovnLogLine, 10)).To(BeNil())
				// Read line from Log Forward output
				raw, err := framework.ReadAuditLogsFrom(logging.OutputTypeElasticsearch)
				Expect(err).To(BeNil(), "Expected no errors reading the logs")
				//var logs []types.OVNAuditLog
				//err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
				//ExpectOK(err)
				// Compare to expected template
				//outputTestLog := logs[0]
				//Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
				results := strings.Join(raw, " ")
				Expect(results).To(MatchRegexp("name=verify-audit-logging_deny-all"), "Message should contain the audit log: %v", raw)

				// Parse logs and verify level and type
				logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
				ExpectOK(err, "Expected no errors parsing the logs: %s", raw)
				Expect(logs[0].Level).To(Equal(outputLogTemplate.Level))
				Expect(logs[0].LogType).To(Equal(outputLogTemplate.LogType))
			})

			AfterEach(func() {
				framework.Cleanup()
			})
		}
	})

})
