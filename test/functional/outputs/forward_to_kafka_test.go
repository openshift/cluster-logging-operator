package outputs

import (
	"github.com/ViaQ/logerr/v2/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
)

var _ = Describe("[Functional][Outputs][Kafka] Functional tests", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()

		log.NewLogger("").V(2).Info("Creating secret for broker credentials")
		brokersecret := kafka.NewBrokerSecret(framework.Namespace)
		if err := framework.Test.Client.Create(brokersecret); err != nil {
			panic(err)
		}

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	//What would be other Kafka specific values ?
	setKafkaSpecValues := func(outspec *logging.OutputSpec) {
		outspec.Kafka = &logging.Kafka{
			Topic: kafka.AppLogsTopic,
		}
	}
	join := func(
		f1 func(spec *logging.OutputSpec),
		f2 func(spec *logging.OutputSpec)) func(*logging.OutputSpec) {
		return func(s *logging.OutputSpec) {
			f1(s)
			f2(s)
		}
	}

	//	timestamp := "2013-03-28T14:36:03.243000+00:00"

	Context("Application Logs", func() {
		It("should send large message over Kafka", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setKafkaSpecValues, func(spec *logging.OutputSpec) {
					//at this port fluent connects with kafka broker
					spec.URL = "https://localhost:9093"
					spec.Secret.Name = "kafka"
				}), logging.OutputTypeKafka)
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
