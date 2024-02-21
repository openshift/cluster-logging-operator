// go:build !fluentd

package azuremonitor

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/azuremonitor"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

const (
	azureSecret   = "azure-secret"
	application   = "application"
	serverStarted = "Server started on port 3000"
	failedReason  = "reason=\"Service call failed. No retries or retries exhausted.\""
)

var _ = Describe("Forwarding to Azure Monitor Log ", func() {
	var (
		framework   *functional.CollectorFunctionalFramework
		sharedKey   = rand.Word(16)
		customerId  = strings.ToLower(string(rand.Word(16)))
		azureDomain = "acme.com"
	)

	BeforeEach(func() {
		azureHostPort := fmt.Sprintf("%s:%d", azureDomain, azuremonitor.Port)

		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		secret := runtime.NewSecret(framework.Namespace, azureSecret,
			map[string][]byte{
				constants.SharedKey: sharedKey,
			},
		)
		framework.Secrets = append(framework.Secrets, secret)

		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInputWithVisitor("custom-app",
				func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{}
				}).
			ToOutputWithVisitor(
				func(spec *logging.OutputSpec) {
					spec.Type = logging.OutputTypeAzureMonitor
					spec.OutputTypeSpec = logging.OutputTypeSpec{
						AzureMonitor: &logging.AzureMonitor{
							LogType:    "myLogType",
							CustomerId: customerId,
							Host:       azureHostPort,
						},
					}
					spec.TLS = &logging.OutputTLSSpec{
						InsecureSkipVerify: true,
					}
					spec.Secret = &logging.OutputSecretSpec{
						Name: azureSecret,
					}
				},
				logging.OutputTypeAzureMonitor)

		Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
			altHost := fmt.Sprintf("%s.%s", customerId, azureDomain)
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
		appLogTemplate.Timestamp = nanoTime
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
