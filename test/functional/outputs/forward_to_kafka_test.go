package outputs

import (
	"github.com/ViaQ/logerr/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
)

var _ = Describe("[LogForwarding][Kafka] Functional tests", func() {

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()

		log.V(2).Info("Creating secret for broker credentials")
		brokersecret := kafka.NewBrokerSecret(framework.Namespace)
		if err := framework.Test.Client.Create(brokersecret); err != nil {
			panic(err)
		}

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("Application Logs", func() {
		It("should send large message over Kafka", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToKafkaOutput()
			Expect(framework.Deploy()).To(BeNil())

			maxLen := uint64(1000)
			Expect(framework.WritesNApplicationLogsOfSize(1, maxLen)).To(BeNil())
			// Read line from Kafka output
			outputlogs, err := framework.ReadApplicationLogsFromKafka("clo-app-topic", "localhost:9092", "kafka-consumer-clo-app-topic")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		})
	})

})
