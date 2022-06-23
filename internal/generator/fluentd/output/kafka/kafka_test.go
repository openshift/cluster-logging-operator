package kafka

import (
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate fluentd config", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Forwarder != nil &&
			clspec.Forwarder.Fluentd != nil &&
			clspec.Forwarder.Fluentd.Buffer != nil {
			bufspec = clspec.Forwarder.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], op)
	}
	DescribeTable("for kafka output", helpers.TestGenerateConfWith(f),
		Entry("with username,password to single topic", helpers.ConfGenerateTest{
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
<label @KAFKA_RECEIVER>
  <match **>
    @type kafka2
    @id kafka_receiver
    brokers broker1-kafka.svc.messaging.cluster.local:9092
    default_topic build_complete
    use_event_time true
    username "#{File.exists?('/var/run/ocp-collector/secrets/kafka-receiver-1/username') ? open('/var/run/ocp-collector/secrets/kafka-receiver-1/username','r') do |f|f.read end : ''}"
    password "#{File.exists?('/var/run/ocp-collector/secrets/kafka-receiver-1/password') ? open('/var/run/ocp-collector/secrets/kafka-receiver-1/password','r') do |f|f.read end : ''}"
    sasl_over_ssl false
    <format>
      @type json
    </format>
    <buffer build_complete>
      @type file
      path '/var/lib/fluentd/kafka_receiver'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
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
<label @KAFKA_RECEIVER>
  <match **>
    @type kafka2
    @id kafka_receiver
    brokers broker1-kafka.svc.messaging.cluster.local:9092
    default_topic topic
    use_event_time true
    ssl_client_cert_key '/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key'
    ssl_client_cert '/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt'
    ssl_ca_cert '/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt'
    sasl_over_ssl false
    <format>
      @type json
    </format>
    <buffer topic>
      @type file
      path '/var/lib/fluentd/kafka_receiver'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("without security", helpers.ConfGenerateTest{
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
<label @KAFKA_RECEIVER>
  <match **>
    @type kafka2
    @id kafka_receiver
    brokers broker1-kafka.svc.messaging.cluster.local:9092
    default_topic topic
    use_event_time true
    <format>
      @type json
    </format>
    <buffer topic>
      @type file
      path '/var/lib/fluentd/kafka_receiver'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
	)
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
