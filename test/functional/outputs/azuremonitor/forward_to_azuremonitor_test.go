package azuremonitor

import (
	"fmt"
	"strings"
	"time"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/azuremonitor"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	failedReason = "reason=\"Service call failed. No retries or retries exhausted.\""
)

var _ = Describe("Forwarding to Azure Monitor Log ", func() {
	var (
		framework  *functional.CollectorFunctionalFramework
		sharedKey  = rand.Word(16)
		customerId = strings.ToLower(string(rand.Word(16)))
	)

	BeforeEach(func() {

		framework = functional.NewCollectorFunctionalFramework()
		secret := runtime.NewSecret(framework.Namespace, azuremonitor.AzureSecretName,
			map[string][]byte{
				constants.SharedKey: sharedKey,
			},
		)
		framework.Secrets = append(framework.Secrets, secret)

		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToAzureMonitorOutput(func(output *obs.OutputSpec) {
				output.AzureMonitor.CustomerId = customerId
			})
		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			altHost := fmt.Sprintf("%s.%s", customerId, azuremonitor.AzureDomain)
			return azuremonitor.NewMockoonVisitor(b, altHost, framework)
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
		var appLogTemplate = functional.NewApplicationLogTemplate()
		appLogTemplate.TimestampLegacy = nanoTime
		appLogTemplate.Message = message
		appLogTemplate.Level = "default"
		appLogTemplate.Kubernetes.PodName = framework.Pod.Name
		appLogTemplate.Kubernetes.ContainerName = constants.CollectorName
		appLogTemplate.Kubernetes.NamespaceName = framework.Namespace

		applicationLogLine := functional.NewCRIOLogMessage(timestamp, message, false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 3)).To(BeNil())
		time.Sleep(30 * time.Second)

		//read log from collector container
		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil())
		Expect(strings.Count(collectorLog, failedReason)).To(BeEquivalentTo(0))

		//read log from mock server container and validate it
		appLogs, err := azuremonitor.ReadApplicationLogFromMockoon(framework)
		Expect(err).To(BeNil())
		Expect(appLogs).ToNot(BeNil())
		Expect(len(appLogs)).To(BeEquivalentTo(3))
		for i := 0; i < 3; i++ {
			Expect(appLogs[i]).To(matchers.FitLogFormatTemplate(appLogTemplate))
		}
	})
})
