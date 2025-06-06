package normalization

import (
	"fmt"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][LogForwarding][Normalization] message format tests for audit logs", func() {
	defer GinkgoRecover()
	var (
		framework *functional.CollectorFunctionalFramework
	)

	Context("with vector", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToElasticSearchOutput()
			Expect(framework.Deploy()).To(BeNil())
		})

		It("should parse k8s audit log format correctly", func() {
			// Log message data
			timestamp := "2013-03-28T14:36:03.243000+00:00"
			nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

			// Define a template for test format (used for input, and expected output)
			var outputLogTemplate = types.K8sAuditLog{
				AuditLogCommon: types.AuditLogCommon{
					Kind:             "Event",
					Hostname:         framework.Pod.Spec.NodeName,
					LogSource:        "*",
					LogType:          "audit",
					Level:            "Metadata",
					Timestamp:        time.Time{},
					TimestampLegacy:  time.Time{},
					PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
					OpenshiftLabels: types.OpenshiftMeta{
						ClusterID: "*",
					},
				},
				K8SAuditLevel: "Metadata",
			}
			// Template expected as output Log
			k8sAuditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata"}`, functional.CRIOTime(nanoTime))
			Expect(framework.WriteMessagesTok8sAuditLog(k8sAuditLogLine, 10)).To(BeNil())
			// Read line from Log Forward output
			raw, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			var logs []types.K8sAuditLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs: %v", raw)
			// Compare to expected template
			outputTestLog := logs[0]
			Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
			results := strings.Join(raw, " ")
			Expect(results).To(MatchRegexp("kind.*Event.*level.*Metadata.*k8s_audit_level.*Metadata"), "Message should contain the audit log: %v", raw)

		})
		It("should parse openshift audit log format correctly", func() {
			// Log message data
			timestamp := "2013-03-28T14:36:03.243000+00:00"
			nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)

			// Define a template for test format (used for input, and expected output)
			var outputLogTemplate = types.OpenshiftAuditLog{
				AuditLogCommon: types.AuditLogCommon{
					Kind:             "Event",
					Hostname:         framework.Pod.Spec.NodeName,
					LogSource:        "*",
					LogType:          "audit",
					Level:            "Metadata",
					Timestamp:        time.Time{},
					TimestampLegacy:  time.Time{},
					PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
					OpenshiftLabels: types.OpenshiftMeta{
						ClusterID: "*",
					},
				},
				OpenshiftAuditLevel: "Metadata",
			}
			// Template expected as output Log
			auditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata"}`, functional.CRIOTime(nanoTime))
			Expect(framework.WriteMessagesToOpenshiftAuditLog(auditLogLine, 10)).To(BeNil())
			Expect(framework.WriteMessagesToOAuthAuditLog(auditLogLine, 10)).To(BeNil())
			// Read line from Log Forward output
			raw, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			var logs []types.OpenshiftAuditLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs: %v", raw)
			// Compare to expected template
			for _, outputTestLog := range logs {
				Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
				results := strings.Join(raw, " ")
				Expect(results).To(MatchRegexp("kind.*Event.*level.*Metadata.*openshift_audit_level.*Metadata"), "Message should contain the audit log: %v", raw)
			}
		})
		It("should parse linux audit log format correctly", func() {
			// Log message data
			timestamp := "2024-03-28T14:36:03.243000+00:00"
			testTime, _ := time.Parse(time.RFC3339Nano, timestamp)
			auditLogLine := functional.NewAuditHostLog(testTime)
			// Template expected as output Log
			var outputLogTemplate = types.LinuxAuditLog{
				Message:   auditLogLine,
				LogSource: "*",
				LogType:   "audit",
				Level:     "default",
				Hostname:  framework.Pod.Spec.NodeName,
				AuditLinux: types.AuditLinux{
					Type:     "DAEMON_START",
					RecordID: "*",
				},
				Timestamp:        testTime,
				TimestampLegacy:  testTime,
				PipelineMetadata: functional.TemplateForAnyPipelineMetadata,
				Openshift: types.OpenshiftMeta{
					ClusterID: "*",
				},
			}
			// Write log line as input to fluentd
			Expect(framework.WriteMessagesToAuditLog(auditLogLine, 10)).To(BeNil())
			// Read line from Log Forward output
			raw, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			var logs []types.LinuxAuditLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputTestLog := logs[0]
			Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
			results := strings.Join(raw, " ")
			Expect(results).To(MatchRegexp("format=enriched kernel="), "Message should contain the audit log: %v", raw)
		})
		It("should parse ovn audit log correctly", func() {
			// Log message data
			level := "info"
			creationTime := time.Now()
			ovnLogLine := functional.NewOVNAuditLog(creationTime)

			// Template expected as output Log
			var outputLogTemplate = types.OVNAuditLog{
				Message:         ovnLogLine,
				Level:           level,
				Hostname:        framework.Pod.Spec.NodeName,
				Timestamp:       time.Time{},
				TimestampLegacy: time.Time{},
				LogSource:       string(obs.AuditSourceOVN),
				LogType:         string(obs.InputTypeAudit),
				Openshift: types.OpenshiftMeta{
					Sequence:  types.NewOptionalInt(""),
					ClusterID: "*",
				},
			}

			Expect(framework.WriteMessagesToOVNAuditLog(ovnLogLine, 10)).To(BeNil())

			// Read line from Log Forward output
			raw, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")

			var logs []types.OVNAuditLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
			ExpectOK(err)

			// Compare to expected template
			outputTestLog := logs[0]
			Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))
			results := strings.Join(raw, " ")
			Expect(results).To(MatchRegexp("name=verify-audit-logging_deny-all"), "Message should contain the audit log: %v", raw)
			Expect(outputTestLog.TimestampLegacy).To(Equal(outputTestLog.Timestamp))
		})

		AfterEach(func() {
			framework.Cleanup()
		})
	})

})
