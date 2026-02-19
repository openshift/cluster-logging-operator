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
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/azure/logsingestion"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	failedReason = "reason=\"Service call failed. No retries or retries exhausted.\""
)

var _ = Describe("Forwarding to Azure Log Ingestion", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		secret := runtime.NewSecret(framework.Namespace, logsingestion.SecretName,
			map[string][]byte{
				logsingestion.ClientSecretKeyName: []byte("fake-client-secret-for-testing"),
			},
		)
		framework.Secrets = append(framework.Secrets, secret)

		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToAzureLogsIngestionOutput()
		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			return logsingestion.NewMockoonVisitor(b, framework)
		})).To(BeNil())
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should accept application logs", func() {
		// Write application logs
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
		time.Sleep(30 * time.Second)

		// Read log from collector container and verify no service call failures
		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil())
		Expect(strings.Count(collectorLog, failedReason)).To(BeEquivalentTo(0))

		// Read log from mock server container and validate it
		appLogs, err := logsingestion.ReadApplicationLog(framework)
		Expect(err).To(BeNil())
		Expect(appLogs).ToNot(BeNil())
		Expect(appLogs).To(HaveLen(3))
		for i := range 3 {
			Expect(appLogs[i]).To(matchers.FitLogFormatTemplate(appLogTemplate))
		}
	})
})
