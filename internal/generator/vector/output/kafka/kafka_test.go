package kafka

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate vector config", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		return New(vectorhelpers.FormatComponentID(clfspec.Outputs[0].Name), clfspec.Outputs[0], []string{"pipeline_1", "pipeline_2"}, secrets[clfspec.Outputs[0].Name], nil, op)
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "build_complete"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "build_complete"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"
`,
		}),
		Entry("brokers, no URL, with tls key,cert,ca-bundle", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						OutputTypeSpec: logging.OutputTypeSpec{
							Kafka: &logging.Kafka{
								Topic:   `topic`,
								Brokers: []string{`tls://broker1:9092`, `tls://broker2:9092`, `tls://broker3:9092`},
							},
						},
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1:9092,broker2:9092,broker3:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"
`,
		}),
		Entry("with TLS and InsecureSkipVerify", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

[sinks.kafka_receiver.librdkafka_options]
"enable.ssl.certificate.verification" = "false"
[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"
`,
		}),
		Entry("with TLS Key Pass", helpers.ConfGenerateTest{
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
						"passphrase": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

[sinks.kafka_receiver.tls]
enabled = true
key_pass = "junk"
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"
`,
		}),
		Entry("brokers, no URL, with tls key,cert,ca-bundle", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						OutputTypeSpec: logging.OutputTypeSpec{
							Kafka: &logging.Kafka{
								Topic:   `topic`,
								Brokers: []string{`tcp://broker1:9092`, `tcp://broker2:9092`, `tcp://broker3:9092`},
							},
						},
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
[transforms.kafka_receiver_dedot]
type = "remap"
inputs = ["pipeline_1","pipeline_2"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["kafka_receiver_dedot"]
bootstrap_servers = "broker1:9092,broker2:9092,broker3:9092"
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
