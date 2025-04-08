package kafka

import (
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
)

var _ = Describe("[Functional][Outputs][Kafka] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		log.V(2).Info("Creating secret for broker credentials")
		framework.Secrets = append(framework.Secrets, kafka.NewBrokerSecret(framework.Namespace))

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("Application Logs", func() {
		It("should send large message over Kafka", func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToKafkaOutput()
			Expect(framework.Deploy()).To(BeNil())

			maxLen := 1000
			Expect(framework.WritesNApplicationLogsOfSize(1, maxLen, 0)).To(BeNil())
			// Read line from Kafka output
			outputlogs, err := framework.ReadApplicationLogsFromKafka("clo-app-topic", "localhost:9092", "kafka-consumer-clo-app-topic")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		})
	})
	Context("LOG-3458", func() {
		It("should deliver message to a topic named the same as payload key", func() {
			topic := "openshift"
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToKafkaOutput(func(output *obs.OutputSpec) {
					output.Kafka.Topic = topic
				})
			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(20, 10, 0)).To(BeNil())
			// Read line from Kafka output
			outputlogs, err := framework.ReadApplicationLogsFromKafka(topic, "localhost:9092", kafka.ConsumerNameForTopic(topic))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		})
	})

	Context("topics", func() {
		DescribeTable("user defined topics", func(topic, expTopic string) {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToKafkaOutput(func(output *obs.OutputSpec) {
					output.Kafka.Topic = topic
				})
			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(20, 10, 0)).To(BeNil())
			// Read line from Kafka output
			outputlogs, err := framework.ReadApplicationLogsFromKafka(expTopic, "localhost:9092", kafka.ConsumerNameForTopic(expTopic))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		},
			Entry("should write to defined static topic", "custom-index", "custom-index"),
			Entry("should write to defined dynamic topic", `{.log_type||"none"}`, "application"),
			Entry("should write to defined static + dynamic topic", `foo-{.log_type||"none"}`, "foo-application"),
			Entry("should write to defined static + fallback value if field is missing", `foo-{.missing||"none"}`, "foo-none"))
	})

	Context("with tuning parameters", func() {

		DescribeTable("with compression", func(compression string) {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToKafkaOutput(func(output *obs.OutputSpec) {
					output.Kafka.Tuning = &obs.KafkaTuningSpec{
						Compression: compression,
					}
				})

			Expect(framework.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			// Read line from Kafka output
			logs, err := framework.ReadApplicationLogsFromKafka("clo-app-topic", "localhost:9092", "kafka-consumer-clo-app-topic")
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", obs.OutputTypeKafka, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", obs.OutputTypeKafka)

		},
			Entry("should pass with snappy", "snappy"),
			Entry("should pass with zstd", "zstd"),
			Entry("should pass with lz4", "lz4"),
			Entry("should pass with none", "none"))
	})

})
