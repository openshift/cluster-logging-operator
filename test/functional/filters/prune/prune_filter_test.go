package prune

import (
	"fmt"
	"strings"
	"time"

	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

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
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
			f = functional.NewCollectorFunctionalFramework()
			specialCharLabel := "foo-bar/baz"
			f.Labels = map[string]string{specialCharLabel: "specialCharLabel"}

			testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithFilter(pruneFilterName, func(spec *obs.FilterSpec) {
					spec.Type = obs.FilterTypePrune
					spec.PruneFilterSpec = &obs.PruneFilterSpec{
						In:    []obs.FieldPath{".kubernetes.namespace_name", ".kubernetes.container_name", `.kubernetes.labels."foo-bar/baz"`},
						NotIn: []obs.FieldPath{".log_type", ".log_source", ".message", ".kubernetes", ".openshift", `."@timestamp"`},
					}
				}).
				ToElasticSearchOutput()

			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "my error message")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			logs, err := f.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", obs.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", obs.OutputTypeElasticsearch)

			log := logs[0]

			Expect(log.Message).ToNot(BeNil())
			Expect(log.LogType).ToNot(BeNil())
			Expect(log.Kubernetes).ToNot(BeNil())
			Expect(log.Openshift).ToNot(BeNil())
			Expect(log.Timestamp).ToNot(BeNil())
			Expect(log.Kubernetes.Annotations).ToNot(BeNil())
			Expect(log.Kubernetes.PodName).ToNot(BeNil())
			Expect(log.Kubernetes.Labels).ToNot(ContainElement("foo-bar_baz"))

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
			secretKey  = internalobs.NewSecretReference("hecToken", "do-not-tell")
		)

		BeforeEach(func() {
			f = functional.NewCollectorFunctionalFramework()
			pipelineBuilder = testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithFilter(pruneFilterName, func(spec *obs.FilterSpec) {
					spec.Type = obs.FilterTypePrune
					spec.PruneFilterSpec = &obs.PruneFilterSpec{NotIn: []obs.FieldPath{".log_type", ".message"}}
				})
		})

		It("should send to Elasticsearch with only .log_type and .message", func() {
			pipelineBuilder.ToElasticSearchOutput()
			Expect(f.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			logs, err := f.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", obs.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", obs.OutputTypeElasticsearch)
		})

		It("should send to Splunk with only .log_type and .message", func() {
			pipelineBuilder.ToSplunkOutput(*secretKey)
			secret = runtime.NewSecret(f.Namespace, secretKey.SecretName,
				map[string][]byte{
					secretKey.Key: functional.HecToken,
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
			pipelineBuilder.ToOutputWithVisitor(func(spec *obs.OutputSpec) {
				spec.Type = obs.OutputTypeLoki
				spec.Loki = &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: l.InternalURL("").String(),
					},
				}
			}, string(obs.OutputTypeLoki))

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
			logs, err := f.ReadApplicationLogsFrom(string(obs.OutputTypeKafka))
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", obs.OutputTypeKafka, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", obs.OutputTypeKafka)
		})

		It("should send to HTTP with only .log_type and .message", func() {
			pipelineBuilder.ToHttpOutput()
			Expect(f.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			raw, err := f.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())
		})

		It("should send to Syslog with only .log_type and .message", func() {
			pipelineBuilder.ToSyslogOutput(obs.SyslogRFC5424)
			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			outputlogs, err := f.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		})

		It("should send to AzureMonitor with only .log_type and .message", func() {
			pipelineBuilder.ToAzureMonitorOutput(func(output *obs.OutputSpec) {
				output.AzureMonitor.CustomerId = customerId
			})

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
			pipelineBuilder.ToCloudwatchOutput(obs.CloudwatchAuthentication{
				Type: obs.CloudwatchAuthTypeAccessKey,
				AWSAccessKey: &obs.CloudwatchAWSAccessKey{
					KeyId: obs.SecretReference{
						Key:        constants.AWSAccessKeyID,
						SecretName: functional.CloudwatchSecret,
					},
					KeySecret: obs.SecretReference{
						Key:        constants.AWSSecretAccessKey,
						SecretName: functional.CloudwatchSecret,
					},
				},
			})

			secret = runtime.NewSecret(f.Namespace, functional.CloudwatchSecret,
				map[string][]byte{
					constants.AWSAccessKeyID:     []byte(functional.AwsAccessKeyID),
					constants.AWSSecretAccessKey: []byte(functional.AwsSecretAccessKey),
				},
			)

			f.Secrets = append(f.Secrets, secret)

			Expect(f.Deploy()).To(BeNil())

			// Write/ read logs
			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			time.Sleep(10 * time.Second)

			logs, err := f.ReadLogsFromCloudwatch(string(obs.InputTypeApplication))
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(1))
		})
	})
})
