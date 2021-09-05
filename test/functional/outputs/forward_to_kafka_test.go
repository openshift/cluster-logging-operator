package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"strings"
	//. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[LogForwarding][Kafka] Functional tests", func() {

	var (
		framework *functional.FluentdFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
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
		FIt("should send large message over Kafka", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(join(setKafkaSpecValues, func(spec *logging.OutputSpec) {
					spec.URL = "http://0.0.0.0:9094"
				}), logging.OutputTypeKafka)
			Expect(framework.Deploy()).To(BeNil())

			var MaxLen uint64 = 40000
			Expect(framework.WritesNApplicationLogsOfSize(1, MaxLen)).To(BeNil())
			// Read line from Kafka output
			outputlogs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeKafka)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).ToNot(BeEmpty())
			fields := strings.Split(outputlogs[0], "#011")
			msg := fields[2]
			// adjust for "message:" prefix in the received message
			ReceivedLen := uint64(len(msg[8:]))
			Expect(ReceivedLen).To(Equal(MaxLen))
		})
	})

})


