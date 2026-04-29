package gcl

import (
	"fmt"
	"strings"
	"time"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	internalgcl "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	gclhelper "github.com/openshift/cluster-logging-operator/test/helpers/gcl"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Forwarding to GCP Cloud Logging", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)

	setup := func(inputType obs.InputType, testOptions ...client.TestOption) {
		framework = functional.NewCollectorFunctionalFramework(testOptions...)

		saJSON, err := gclhelper.GenerateFakeServiceAccountJSON(
			fmt.Sprintf("https://%s/token", gclhelper.GCLDomain))
		Expect(err).To(BeNil())

		secret := runtime.NewSecret(framework.Namespace, gclhelper.SecretName,
			map[string][]byte{
				internalgcl.GoogleApplicationCredentialsKey: saJSON,
			},
		)
		framework.Secrets = append(framework.Secrets, secret)

		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(inputType).
			ToGoogleCloudLoggingOutput()
		Expect(framework.DeployWithVisitor(
			gclhelper.NewMockoonVisitor(framework),
		)).To(BeNil())
	}

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("application logs", func() {
		BeforeEach(func() {
			setup(obs.InputTypeApplication)
		})

		It("should accept application logs", func() {
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
			message := "This is my new test message"
			appLogTemplate := functional.NewApplicationLogTemplate()
			appLogTemplate.TimestampLegacy = nanoTime
			appLogTemplate.Message = message
			appLogTemplate.Level = "**optional**"
			appLogTemplate.Kubernetes.PodName = framework.Pod.Name
			appLogTemplate.Kubernetes.ContainerName = constants.CollectorName
			appLogTemplate.Kubernetes.NamespaceName = framework.Namespace

			applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 3)).To(BeNil())
			time.Sleep(30 * time.Second)

			collectorLog, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLog).ToNot(ContainSubstring("error sending request"))

			appLogs, err := gclhelper.ReadApplicationLog(framework.Namespace, framework.Name)
			Expect(err).To(BeNil())
			Expect(appLogs).ToNot(BeNil())
			Expect(appLogs).To(HaveLen(3))
			for i := range 3 {
				Expect(appLogs[i]).To(matchers.FitLogFormatTemplate(appLogTemplate))
			}
		})
	})

	Context("infrastructure logs", func() {
		BeforeEach(func() {
			setup(obs.InputTypeInfrastructure, client.UseInfraNamespaceTestOption)
		})

		It("should accept infrastructure container logs", func() {
			infraLogTemplate := functional.NewContainerInfrastructureLogTemplate()
			infraLogTemplate.Level = "**optional**"
			infraLogTemplate.Kubernetes.ContainerStream = "**optional**"

			Expect(framework.WritesInfraContainerLogs(3)).To(BeNil())
			time.Sleep(30 * time.Second)

			collectorLog, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLog).ToNot(ContainSubstring("error sending request"))

			infraLogs, err := gclhelper.ReadApplicationLog(framework.Namespace, framework.Name)
			Expect(err).To(BeNil())
			Expect(infraLogs).ToNot(BeNil())
			Expect(infraLogs).To(HaveLen(3))
			for i := range 3 {
				Expect(infraLogs[i]).To(matchers.FitLogFormatTemplate(infraLogTemplate))
			}
		})
	})

	Context("audit logs", func() {
		BeforeEach(func() {
			setup(obs.InputTypeAudit)
		})

		It("should accept audit logs", func() {
			Expect(framework.WriteK8sAuditLog(1)).To(BeNil())
			time.Sleep(30 * time.Second)

			collectorLog, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLog).ToNot(ContainSubstring("error sending request"))

			rawEntries, err := gclhelper.ReadRawLogEntries(framework.Namespace, framework.Name)
			Expect(err).To(BeNil())
			Expect(rawEntries).ToNot(BeEmpty())

			var auditLogs []types.AuditLogCommon
			jsonString := fmt.Sprintf("[%s]", strings.Join(rawEntries, ","))
			Expect(types.ParseLogsFrom(jsonString, &auditLogs, false)).To(Succeed())
			Expect(auditLogs).To(HaveLen(1))
			Expect(auditLogs[0].LogType).To(Equal(string(obs.InputTypeAudit)))
			Expect(auditLogs[0].Timestamp).ToNot(BeZero())
		})
	})
})
