package prune

import (
	"fmt"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/azuremonitor"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

var _ = Describe("[Functional][Filters][Prune] Prune filter", func() {
	const (
		pruneFilterName = "my-prune"
	)

	var (
		f *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		f.Cleanup()
	})

	Describe("when prune filter is spec'd", func() {
		It("should prune logs of fields not defined in `NotIn` first and then prune fields defined in `In`", func() {
			f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
			specialCharLabel := "foo-bar/baz"
			f.Labels = map[string]string{specialCharLabel: "specialCharLabel"}

			testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(logging.InputNameApplication).
				WithFilterWithVisitor(pruneFilterName, func(spec *logging.FilterSpec) {
					spec.Type = logging.FilterPrune
					spec.FilterTypeSpec = logging.FilterTypeSpec{
						PruneFilterSpec: &logging.PruneFilterSpec{
							In:    []string{".kubernetes.namespace_name", ".kubernetes.container_name", `.kubernetes.labels."foo-bar/baz"`},
							NotIn: []string{".log_type", ".message", ".kubernetes", ".openshift", `."@timestamp"`},
						},
					}
				}).
				ToElasticSearchOutput()

			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "my error message")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			logs, err := f.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeElasticsearch)

			log := logs[0]

			Expect(log.Message).ToNot(BeNil())
			Expect(log.LogType).ToNot(BeNil())
			Expect(log.Kubernetes).ToNot(BeNil())
			Expect(log.Openshift).ToNot(BeNil())
			Expect(log.Timestamp).ToNot(BeNil())
			Expect(log.Kubernetes.Annotations).ToNot(BeNil())
			Expect(log.Kubernetes.PodName).ToNot(BeNil())
			Expect(log.Kubernetes.FlatLabels).To(HaveLen(1))
			Expect(log.Kubernetes.FlatLabels).ToNot(ContainElement("foo-bar_baz=specialCharLabel"))

			Expect(log.Kubernetes.ContainerName).To(BeEmpty())
			Expect(log.Kubernetes.NamespaceName).To(BeEmpty())
			Expect(log.Level).To(BeEmpty())

		})
	})

	Context("minimal set of fields for each output", func() {
		var (
			pipelineBuilder *testruntime.PipelineBuilder
			secret          *v1.Secret

			sharedKey  = rand.Word(16)
			customerId = strings.ToLower(string(rand.Word(16)))
		)

		BeforeEach(func() {
			f = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
			pipelineBuilder = testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(logging.InputNameApplication).
				WithFilterWithVisitor(pruneFilterName, func(spec *logging.FilterSpec) {
					spec.Type = logging.FilterPrune
					spec.FilterTypeSpec = logging.FilterTypeSpec{
						PruneFilterSpec: &logging.PruneFilterSpec{NotIn: []string{".log_type", ".message"}},
					}
				})
		})

		It("should send to Elasticsearch with only .log_type and .message", func() {
			pipelineBuilder.ToElasticSearchOutput()
			Expect(f.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			logs, err := f.ReadApplicationLogsFrom(logging.OutputTypeElasticsearch)
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeElasticsearch)
		})

		It("should send to Splunk with only .log_type and .message", func() {
			pipelineBuilder.ToSplunkOutput()
			secret = runtime.NewSecret(f.Namespace, "splunk-secret",
				map[string][]byte{
					"hecToken": []byte(string(functional.HecToken)),
				},
			)
			f.Secrets = append(f.Secrets, secret)

			Expect(f.Deploy()).To(BeNil())
			time.Sleep(90 * time.Second)

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			// Get logs
			logs, err := f.ReadAppLogsByIndexFromSplunk(f.Namespace, f.Name, "*")
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())
		})

		It("should send to Loki with only .log_type and .message", func() {
			l := loki.NewReceiver(f.Namespace, "loki-server")
			Expect(l.Create(f.Test.Client)).To(Succeed())
			pipelineBuilder.ToOutputWithVisitor(func(spec *logging.OutputSpec) {
				spec.Type = logging.OutputTypeLoki
				spec.URL = l.InternalURL("").String()
			}, logging.OutputTypeLoki)

			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			_, err := l.QueryUntil(`{log_type=~".+"}`, "", 1)
			Expect(err).To(Succeed())
		})

		It("should send to Kafka with only .log_type and .message", func() {
			pipelineBuilder.ToKafkaOutput()
			f.Secrets = append(f.Secrets, kafka.NewBrokerSecret(f.Namespace))

			Expect(f.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			// Read line from Kafka output
			logs, err := f.ReadApplicationLogsFrom(logging.OutputTypeKafka)
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeKafka, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeKafka)
		})

		It("should send to HTTP with only .log_type and .message", func() {
			pipelineBuilder.ToHttpOutput()
			Expect(f.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			raw, err := f.ReadRawApplicationLogsFrom(logging.OutputTypeHttp)
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())
		})

		It("should send to Syslog with only .log_type and .message", func() {
			pipelineBuilder.ToSyslogOutput()
			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			outputlogs, err := f.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		})

		It("should send to AzureMonitor with only .log_type and .message", func() {
			pipelineBuilder.ToAzureMonitorOutputWithCuId(customerId)

			secret := runtime.NewSecret(f.Namespace, azuremonitor.AzureSecretName,
				map[string][]byte{
					constants.SharedKey: sharedKey,
				},
			)
			f.Secrets = append(f.Secrets, secret)

			Expect(f.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				altHost := fmt.Sprintf("%s.%s", customerId, azuremonitor.AzureDomain)
				return azuremonitor.NewMockoonVisitor(b, altHost, f)
			})).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			time.Sleep(30 * time.Second)
			appLogs, err := azuremonitor.ReadApplicationLogFromMockoon(f)
			Expect(err).To(BeNil())
			Expect(appLogs).ToNot(BeNil())
		})

		It("should send to CloudWatch with only .log_type and .message", func() {
			pipelineBuilder.ToCloudwatchOutput()

			secret = runtime.NewSecret(f.Namespace, functional.CloudwatchSecret,
				map[string][]byte{
					"aws_access_key_id":     []byte(functional.AwsAccessKeyID),
					"aws_secret_access_key": []byte(functional.AwsSecretAccessKey),
				},
			)

			f.Secrets = append(f.Secrets, secret)

			Expect(f.Deploy()).To(BeNil())

			// Write/ read logs
			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			time.Sleep(10 * time.Second)

			logs, err := f.ReadLogsFromCloudwatch(logging.InputNameApplication)
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(1))
		})
	})
})
