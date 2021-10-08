package kafka

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate vector config", func() {
	inputPipeline := []string{"transform_application"}
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Forwarder != nil &&
			clspec.Forwarder.Fluentd != nil &&
			clspec.Forwarder.Fluentd.Buffer != nil {
			bufspec = clspec.Forwarder.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], op, inputPipeline)
	}
	DescribeTable("for kafka output testing in kafka_test.go", generator.TestGenerateConfWith(f),
		Entry("with username,password to single topic 123", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
						OutputTypeSpec: logging.OutputTypeSpec{
							Kafka: &logging.Kafka{
								Topic: "build_complete",
							},
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"username": []byte("junk"),
						"password": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
[sinks.kafka_receiver]
  type = "kafka"
  input = ["transform_application"]
  bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
  topic = "build_complete"
  sasl.username = ""
sasl.password = ""
sasl.enabled = false
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Kafka Conf Generation")
}
