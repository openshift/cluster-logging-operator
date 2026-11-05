package inputs

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func verifyAuditTruncation(actualMessage, expectedOriginal string, maxSize *resource.Quantity, expectTruncated bool) {
	const marker = "..TRUNCATED"
	if expectTruncated {
		Expect(actualMessage).To(ContainSubstring(marker), "Expected truncation marker in audit message")
		Expect(len(actualMessage)).To(BeNumerically("<=", int(maxSize.Value())+len(marker)),
			"Truncated audit message size exceeds configured limit + marker length")
	} else {
		Expect(actualMessage).NotTo(ContainSubstring(marker), "Audit message should not contain truncation marker")
		Expect(actualMessage).To(Equal(expectedOriginal), "Audit message content does not match expected payload")
	}
}

var _ = Describe("[Functional][LogForwarding][Inputs] Audit Log Truncation", func() {
	var framework *functional.CollectorFunctionalFramework

	AfterEach(func() {
		if framework != nil {
			framework.Cleanup()
			framework = nil
		}
	})

	Context("Auditd Source with MaxMessageSize", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
		})

		DescribeTable("auditd message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeAudit, func(spec *obs.InputSpec) {
						spec.Name = "audit-auditd-input"
						spec.Audit = &obs.Audit{
							Sources: []obs.AuditSource{obs.AuditSourceAuditd},
						}
						if maxSize != nil {
							spec.Audit.Tuning = &obs.AuditInputTuningSpec{
								MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				message := strings.Repeat("x", messageSize)
				Expect(framework.WriteMessagesToAuditLog(message, 1)).To(Succeed())

				logLines, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading audit logs")
				Expect(logLines).To(HaveLen(1), "Expected exactly one audit log record")

				verifyAuditTruncation(logLines[0], message, maxSize, expectTruncated)
			},
			Entry("with 2KB limit, 4KB message (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message",
				nil, 512, false),
		)
	})

	Context("Kubernetes API Audit Source with MaxMessageSize", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
		})

		DescribeTable("kubeAPI audit message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeAudit, func(spec *obs.InputSpec) {
						spec.Name = "audit-kube-input"
						spec.Audit = &obs.Audit{
							Sources: []obs.AuditSource{obs.AuditSourceKube},
						}
						if maxSize != nil {
							spec.Audit.Tuning = &obs.AuditInputTuningSpec{
								MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				message := strings.Repeat("x", messageSize)
				Expect(framework.WriteMessagesTok8sAuditLog(message, 1)).To(Succeed())

				logLines, err := framework.Readk8sAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading k8s audit logs")
				Expect(logLines).To(HaveLen(1), "Expected exactly one k8s audit log record")

				verifyAuditTruncation(logLines[0], message, maxSize, expectTruncated)
			},
			Entry("with 2KB limit, 4KB message (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message",
				nil, 512, false),
		)
	})

	Context("OpenShift API Audit Source with MaxMessageSize", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
		})

		DescribeTable("openshiftAPI audit message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeAudit, func(spec *obs.InputSpec) {
						spec.Name = "audit-openshift-input"
						spec.Audit = &obs.Audit{
							Sources: []obs.AuditSource{obs.AuditSourceOpenShift},
						}
						if maxSize != nil {
							spec.Audit.Tuning = &obs.AuditInputTuningSpec{
								MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				message := strings.Repeat("x", messageSize)
				Expect(framework.WriteMessagesToOpenshiftAuditLog(message, 1)).To(Succeed())

				logLines, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading openshift audit logs")
				Expect(logLines).To(HaveLen(1), "Expected exactly one openshift audit log record")

				verifyAuditTruncation(logLines[0], message, maxSize, expectTruncated)
			},
			Entry("with 2KB limit, 4KB message (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message",
				nil, 512, false),
		)
	})

	Context("OVN Audit Source with MaxMessageSize", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
		})

		DescribeTable("OVN audit message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeAudit, func(spec *obs.InputSpec) {
						spec.Name = "audit-ovn-input"
						spec.Audit = &obs.Audit{
							Sources: []obs.AuditSource{obs.AuditSourceOVN},
						}
						if maxSize != nil {
							spec.Audit.Tuning = &obs.AuditInputTuningSpec{
								MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				message := strings.Repeat("x", messageSize)
				Expect(framework.WriteMessagesToOVNAuditLog(message, 1)).To(Succeed())

				logLines, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading OVN audit logs")
				Expect(logLines).To(HaveLen(1), "Expected exactly one OVN audit log record")

				verifyAuditTruncation(logLines[0], message, maxSize, expectTruncated)
			},
			Entry("with 2KB limit, 4KB message (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message",
				nil, 512, false),
		)
	})

	Context("Audit Sources with combined tuning options", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
		})

		DescribeTable("audit message limits with IgnoreOlder and MaxMessageSize",
			func(auditSource obs.AuditSource, maxSize *resource.Quantity, ignoreOlder time.Duration, messageSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeAudit, func(spec *obs.InputSpec) {
						spec.Name = "audit-combined-input"
						spec.Audit = &obs.Audit{
							Sources: []obs.AuditSource{auditSource},
						}
						if maxSize != nil || ignoreOlder > 0 {
							spec.Audit.Tuning = &obs.AuditInputTuningSpec{}
							if maxSize != nil {
								spec.Audit.Tuning.MaxMessageSize = utils.GetPtr(maxSize.DeepCopy())
							}
							if ignoreOlder > 0 {
								spec.Audit.Tuning.IgnoreOlder = &metav1.Duration{Duration: ignoreOlder}
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				message := strings.Repeat("x", messageSize)
				Expect(framework.WriteMessagesToAuditLog(message, 1)).To(Succeed())

				logLines, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading audit logs")
				Expect(logLines).To(HaveLen(1), "Expected exactly one audit log record")

				verifyAuditTruncation(logLines[0], message, maxSize, expectTruncated)
			},
			Entry("auditd with 1KB limit and 1h ignoreOlder, 2KB message (truncated)",
				obs.AuditSourceAuditd, resource.NewQuantity(1024, resource.BinarySI), time.Hour, 2048, true),
			Entry("kubeAPI with 8KB limit and 30m ignoreOlder, 4KB message (not truncated)",
				obs.AuditSourceKube, resource.NewQuantity(8192, resource.BinarySI), 30*time.Minute, 4096, false),
			Entry("openshiftAPI with 2KB limit and 2h ignoreOlder, 3KB message (truncated)",
				obs.AuditSourceOpenShift, resource.NewQuantity(2048, resource.BinarySI), 2*time.Hour, 3072, true),
			Entry("ovn with 6KB limit and 1h ignoreOlder, 4KB message (not truncated)",
				obs.AuditSourceOVN, resource.NewQuantity(6144, resource.BinarySI), time.Hour, 4096, false),
		)
	})
})
