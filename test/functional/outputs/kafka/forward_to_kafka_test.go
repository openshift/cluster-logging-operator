package kafka

import (
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
)

var _ = Describe("[Functional][Outputs][Kafka] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		log.V(2).Info("Creating secret for broker credentials")
		framework.Secrets = append(framework.Secrets, kafka.NewBrokerSecret(framework.Namespace))

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("Application Logs", func() {
		It("should send large message over Kafka", func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
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
				FromInput(logging.InputNameApplication).
				ToKafkaOutput(func(output *logging.OutputSpec) {
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

	Context("with tuning parameters", func() {
		var (
			compVisitFunc func(spec *logging.OutputSpec)
		)
		DescribeTable("with compression", func(compression string) {
			compVisitFunc = func(spec *logging.OutputSpec) {
				spec.Tuning = &logging.OutputTuningSpec{
					Compression: compression,
				}
			}
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToKafkaOutput(compVisitFunc)

			Expect(framework.Deploy()).To(BeNil())

			msg := functional.NewCRIOLogMessage(functional.CRIOTime(time.Now()), "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			// Read line from Kafka output
			logs, err := framework.ReadApplicationLogsFromKafka("clo-app-topic", "localhost:9092", "kafka-consumer-clo-app-topic")
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", logging.OutputTypeKafka, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", logging.OutputTypeKafka)

		},
			Entry("should pass with snappy", "snappy"),
			Entry("should pass with zstd", "zstd"),
			Entry("should pass with lz4", "lz4"),
			Entry("should pass with none", ""))
	})

})
