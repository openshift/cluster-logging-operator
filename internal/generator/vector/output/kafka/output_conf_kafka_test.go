package kafka_test

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	//"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/kafka"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("Generating external kafka server output store config", func() {
	var (
		outputs []v1.OutputSpec
		g       generator.Generator
	)
	BeforeEach(func() {
		g = generator.MakeGenerator()
	})

	Context("for a single kafka default output target out", func() {
		kafkaConf := `
    [sinks.kafka_receiver]
    type = "kafka"
    input = ["transform_application"]
    bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
    topic = "topic"
`

		It("should result in a valid kafka config", func() {
			inputPipeline := []string{"transform_application"}
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
				},
			}

			results, err := g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil, inputPipeline)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
	})
})
