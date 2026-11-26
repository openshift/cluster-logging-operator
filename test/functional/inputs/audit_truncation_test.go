package inputs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/api/resource"
)

const truncationMarker = "..TRUNCATED"

// verifyAuditTruncation validates the log messages against truncation rules.
func verifyAuditTruncation(logLines []string, expectedOriginal string, logSource obs.AuditSource, maxSize *resource.Quantity, expectTruncated bool) {
	Expect(logLines).ToNot(BeEmpty(), "Expected at least one log line to verify")

	actualMessage := logLines[0]
	var logEntry map[string]any
	err := json.Unmarshal([]byte(actualMessage), &logEntry)
	Expect(err).To(BeNil(), "Expected valid JSON in log entry")

	Expect(logEntry["log_type"]).To(Equal(obs.InputTypeAudit.String()), "Expected log_type field to be of type audit")
	Expect(logEntry["log_source"]).To(Equal(logSource.String()), "Expected log_source field to match source type")

	if expectTruncated {
		messageField, ok := logEntry["message"].(string)
		Expect(ok).To(BeTrue(), "Expected 'message' field")
		Expect(messageField).To(ContainSubstring(truncationMarker), "Expected truncation marker in message field")

		if maxSize != nil {
			Expect(len(messageField)).To(BeNumerically("<=", int(maxSize.Value())+len(truncationMarker)),
				"Truncated message size exceeds configured limit + marker length")
		}
		return
	}

	// Handle non-truncated validation
	switch logSource {
	case obs.AuditSourceAuditd, obs.AuditSourceOVN:
		messageField, ok := logEntry["message"].(string)
		Expect(ok).To(BeTrue(), "Expected 'message' field")
		Expect(messageField).NotTo(ContainSubstring(truncationMarker), "Message field should not contain truncation marker")
		Expect(messageField).To(Equal(expectedOriginal), "Message field content does not match expected payload")

	case obs.AuditSourceKube:
		var auditLogs []types.K8sAuditLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logLines, ","))
		err = types.ParseLogsFrom(jsonString, &auditLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the k8s logs")
		Expect(auditLogs).ToNot(BeEmpty())

	case obs.AuditSourceOpenShift:
		var auditLogs []types.OpenshiftAuditLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logLines, ","))
		err = types.ParseLogsFrom(jsonString, &auditLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the openshift logs")
		Expect(auditLogs).ToNot(BeEmpty())
	}
}

func runAuditTruncationTest(
	framework *functional.CollectorFunctionalFramework,
	logSource obs.AuditSource,
	maxSize *resource.Quantity,
	messageSize int,
	expectTruncated bool,
	generateAndWriteLog func(*functional.CollectorFunctionalFramework, int) string,
) {
	testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
		FromInput(obs.InputTypeAudit, func(spec *obs.InputSpec) {
			spec.Name = fmt.Sprintf("audit-%s-input", strings.ToLower(logSource.String()))
			spec.Audit = &obs.Audit{
				Sources: []obs.AuditSource{logSource},
			}
			if maxSize != nil {
				spec.Audit.Tuning = &obs.AuditInputTuningSpec{
					MaxMessageSize: new(maxSize.DeepCopy()),
				}
			}
		}).ToHttpOutput()

	Expect(framework.Deploy()).To(Succeed())

	expectedLog := generateAndWriteLog(framework, messageSize)

	logLines, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeHTTP))
	Expect(err).To(BeNil(), "Expected no errors reading audit logs")
	Expect(logLines).To(HaveLen(1), "Expected exactly one audit log record")

	verifyAuditTruncation(logLines, expectedLog, logSource, maxSize, expectTruncated)
}

var _ = Describe("[Functional][LogForwarding][Inputs] Audit Log Truncation", func() {
	var framework *functional.CollectorFunctionalFramework

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
	})

	AfterEach(func() {
		if framework != nil {
			framework.Cleanup()
			framework = nil
		}
	})

	Context("Auditd Source with MaxMessageSize", func() {
		DescribeTable("auditd message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				runAuditTruncationTest(framework, obs.AuditSourceAuditd, maxSize, messageSize, expectTruncated, func(fw *functional.CollectorFunctionalFramework, size int) string {
					message := strings.Repeat("x", size)
					Expect(fw.WriteMessagesToAuditLog(message, 1)).To(Succeed())
					return message
				})
			},
			Entry("with 2KB limit, 4KB message (truncated)", resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message", nil, 512, false),
		)
	})

	Context("Kubernetes API Audit Source with MaxMessageSize", func() {
		DescribeTable("kubeAPI audit message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				runAuditTruncationTest(framework, obs.AuditSourceKube, maxSize, messageSize, expectTruncated, func(fw *functional.CollectorFunctionalFramework, size int) string {
					k8AuditLog := functional.NewKubeAuditLog(time.Now())
					k8AuditLog = strings.ReplaceAll(k8AuditLog, "ResponseComplete", strings.Repeat("x", size))
					Expect(fw.WriteMessagesTok8sAuditLog(k8AuditLog, 1)).To(Succeed())
					return k8AuditLog
				})
			},
			Entry("with 2KB limit, 4KB message (truncated)", resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message", nil, 512, false),
		)
	})

	Context("OpenShift API Audit Source with MaxMessageSize", func() {
		DescribeTable("openshiftAPI audit message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				runAuditTruncationTest(framework, obs.AuditSourceOpenShift, maxSize, messageSize, expectTruncated, func(fw *functional.CollectorFunctionalFramework, size int) string {
					nowCrio := functional.CRIOTime(time.Now())
					openshiftAuditLogs := fmt.Sprintf(functional.OpenShiftAuditLogTemplate, nowCrio, nowCrio)
					openshiftAuditLogs = strings.ReplaceAll(openshiftAuditLogs, "ResponseComplete", strings.Repeat("x", size))
					Expect(fw.WriteMessagesToOpenshiftAuditLog(openshiftAuditLogs, 1)).To(Succeed())
					return openshiftAuditLogs
				})
			},
			Entry("with 2KB limit, 4KB message (truncated)", resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message", nil, 512, false),
		)
	})

	Context("OVN Audit Source with MaxMessageSize", func() {
		DescribeTable("OVN audit message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				runAuditTruncationTest(framework, obs.AuditSourceOVN, maxSize, messageSize, expectTruncated, func(fw *functional.CollectorFunctionalFramework, size int) string {
					ovnLog := functional.NewOVNAuditLog(time.Now())
					ovnLog = strings.ReplaceAll(ovnLog, "drop", strings.Repeat("x", size))
					Expect(fw.WriteMessagesToOVNAuditLog(ovnLog, 1)).To(Succeed())
					return ovnLog
				})
			},
			Entry("with 2KB limit, 4KB message (truncated)", resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message", nil, 512, false),
		)
	})
})
