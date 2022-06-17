package kafka

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate vector config", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], []string{"pipeline_1", "pipeline_2"}, secrets[clfspec.Outputs[0].Name], op)
	}
	DescribeTable("for kafka output", helpers.TestGenerateConfWith(f),
		Entry("with plaintext sasl, to single topic", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tcp://broker1-kafka.svc.messaging.cluster.local:9092/topic",
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
						"sasl.enable": []byte("true"),
						"username":    []byte("testuser"),
						"password":    []byte("testpass"),
					},
				},
			},
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "build_complete"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"


# SASL Config
[sinks.kafka_receiver.sasl]
enabled = true
username = "testuser"
password = "testpass"
mechanism = "PLAIN"
`,
		}),
		Entry("with tls sasl, to single topic", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "https://broker1-kafka.svc.messaging.cluster.local:9092/topic",
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
						"tls.key":     []byte("junk"),
						"tls.crt":     []byte("junk"),
						"sasl.enable": []byte("true"),
						"username":    []byte("testuser"),
						"password":    []byte("testpass"),
					},
				},
			},
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "build_complete"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

# TLS Config
[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"

# SASL Config
[sinks.kafka_receiver.sasl]
enabled = true
username = "testuser"
password = "testpass"
mechanism = "PLAIN"
`,
		}),
		Entry("with tls sasl, with SCRAM-SHA-256 mechanism to single topic", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "https://broker1-kafka.svc.messaging.cluster.local:9092/topic",
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
						"tls.key":         []byte("junk"),
						"tls.crt":         []byte("junk"),
						"sasl.enable":     []byte("true"),
						"sasl.mechanisms": []byte("SCRAM-SHA-256"),
						"username":        []byte("testuser"),
						"password":        []byte("testpass"),
					},
				},
			},
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "build_complete"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

# TLS Config
[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"

# SASL Config
[sinks.kafka_receiver.sasl]
enabled = true
username = "testuser"
password = "testpass"
mechanism = "SCRAM-SHA-256"
`,
		}),
		Entry("with tls key,cert,ca-bundle", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

# TLS Config
[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"
`,
		}),
		Entry("with TLS and tls.insecure", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
						"tls.insecure":  []byte(""),
					},
				},
			},
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

# TLS Config
[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"

verify_certificate = false
verify_hostname = false
`,
		}),
		Entry("with basic TLS", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"
`,
		}),
		Entry("with plain TLS - no secret", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"
`,
		}),
		Entry("without security", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tcp://broker1-kafka.svc.messaging.cluster.local:9092/topic",
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline_1","pipeline_2"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
