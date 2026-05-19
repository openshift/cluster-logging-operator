package azurelogsingestion

import (
	"strings"
	"time"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/azure/logsingestion"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	failedReason = "reason=\"Service call failed. No retries or retries exhausted.\""
)

var _ = Describe("Forwarding to Azure Log Ingestion", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)

	setupFramework := func(inputType obs.InputType, testOptions ...client.TestOption) {
		framework = functional.NewCollectorFunctionalFramework(testOptions...)
		secret := runtime.NewSecret(framework.Namespace, logsingestion.SecretName,
			map[string][]byte{
				logsingestion.ClientSecretKeyName: []byte("fake-client-secret-for-testing"),
			},
		)
		framework.Secrets = append(framework.Secrets, secret)

		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(inputType).
			ToAzureLogsIngestionOutput()
		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			return logsingestion.NewMockoonVisitor(b, framework)
		})).To(BeNil())
	}

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("with application logs", func() {
		BeforeEach(func() {
			setupFramework(obs.InputTypeApplication)
		})

		It("should accept application logs", func() {
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
			message := "This is my new test message"
			appLogTemplate := functional.NewApplicationLogTemplate()
			appLogTemplate.TimestampLegacy = nanoTime
			appLogTemplate.Message = message
			appLogTemplate.Level = "default"
			appLogTemplate.Kubernetes.PodName = framework.Pod.Name
			appLogTemplate.Kubernetes.ContainerName = constants.CollectorName
			appLogTemplate.Kubernetes.NamespaceName = framework.Namespace

			applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 3)).To(BeNil())

			var appLogs []types.ApplicationLog
			Eventually(func(g Gomega) {
				collectorLog, err := framework.ReadCollectorLogs()
				g.Expect(err).To(BeNil())
				g.Expect(strings.Count(collectorLog, failedReason)).To(BeEquivalentTo(0))

				appLogs, err = logsingestion.ReadApplicationLog(framework)
				g.Expect(err).To(BeNil())
				g.Expect(appLogs).To(HaveLen(3))
			}, 30*time.Second, time.Second).Should(Succeed())
			for i := range 3 {
				Expect(appLogs[i]).To(matchers.FitLogFormatTemplate(appLogTemplate))
			}
		})
	})

	Context("with infrastructure logs", func() {
		BeforeEach(func() {
			setupFramework(obs.InputTypeInfrastructure, client.UseInfraNamespaceTestOption)
		})

		It("should accept infrastructure logs", func() {
			Expect(framework.WritesInfraContainerLogs(3)).To(BeNil())

			Eventually(func(g Gomega) {
				collectorLog, err := framework.ReadCollectorLogs()
				g.Expect(err).To(BeNil())
				g.Expect(strings.Count(collectorLog, failedReason)).To(BeEquivalentTo(0))

				infraLogs, err := logsingestion.ReadInfraLog(framework)
				g.Expect(err).To(BeNil())
				g.Expect(infraLogs).To(HaveLen(3))
				for i := range 3 {
					g.Expect(infraLogs[i].LogType).To(Equal("infrastructure"))
				}
			}, 30*time.Second, time.Second).Should(Succeed())
		})
	})

	Context("with audit logs", func() {
		BeforeEach(func() {
			setupFramework(obs.InputTypeAudit)
		})

		It("should accept kubernetes audit logs", func() {
			Expect(framework.WriteK8sAuditLog(3)).To(BeNil())

			var auditLogs []types.AuditLogCommon
			var rawLogs []map[string]interface{}
			Eventually(func(g Gomega) {
				collectorLog, err := framework.ReadCollectorLogs()
				g.Expect(err).To(BeNil())
				g.Expect(strings.Count(collectorLog, failedReason)).To(BeEquivalentTo(0))

				auditLogs, err = logsingestion.ReadAuditLog(framework)
				g.Expect(err).To(BeNil())
				g.Expect(auditLogs).To(HaveLen(3))

				rawLogs, err = logsingestion.ReadRawLogs(framework)
				g.Expect(err).To(BeNil())
				g.Expect(rawLogs).To(HaveLen(3))
			}, 30*time.Second, time.Second).Should(Succeed())

			for i := range 3 {
				Expect(auditLogs[i].LogType).To(Equal("audit"))
				Expect(auditLogs[i].AuditID).ToNot(BeEmpty())
				Expect(rawLogs[i]).ToNot(HaveKey("kind"), "kind should be renamed to openshift_kind")
				Expect(rawLogs[i]).To(HaveKeyWithValue("openshift_kind", "Event"))
			}
		})
	})
})
