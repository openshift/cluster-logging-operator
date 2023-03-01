package kafka

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
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

	//	timestamp := "2013-03-28T14:36:03.243000+00:00"

	Context("Application Logs", func() {
		It("should send large message over Kafka", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToKafkaOutput()
			Expect(framework.Deploy()).To(BeNil())

			maxLen := 1000
			Expect(framework.WritesNApplicationLogsOfSize(1, maxLen)).To(BeNil())
			// Read line from Kafka output
			outputlogs, err := framework.ReadApplicationLogsFromKafka("clo-app-topic", "localhost:9092", "kafka-consumer-clo-app-topic")
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
		})
	})

})
