package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	v1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var _ = Describe("Generating external kafka server output store config block", func() {
	var (
		err           error
		outputs       []v1.OutputSpec
		forwarderSpec *v1.ForwarderSpec
		generator     *ConfigGenerator
	)
	BeforeEach(func() {
		generator, err = NewConfigGenerator(false, false, false)
		Expect(err).To(BeNil())
	})

	Context("for a single kafka default output target", func() {
		kafkaConf := `<label @KAFKA_RECEIVER>
        <match **>
           @type kafka2
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
               flush_at_shutdown true
               retry_type exponential_backoff
               retry_wait 1s
               retry_max_interval 60s
               retry_forever true
               queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
               total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
               chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
               overflow_action block
           </buffer>
        </match>
        </label>`

		It("should result in a valid kafka label config", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
				},
			}

			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})

		It("should use the default topic if none provided", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092",
				},
			}

			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})
	})

	Context("for kafka output with secured communication and authentication", func() {
		kafkaConf := `<label @KAFKA_RECEIVER>
        <match **>
           @type kafka2
           brokers broker1-kafka.svc.messaging.cluster.local:9092
           default_topic topic
           use_event_time true
           ssl_ca_cert '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
           ssl_client_cert "#{File.exist?('/var/run/ocp-collector/secrets/some-secret/tls.crt') ? '/var/run/ocp-collector/secrets/some-secret/tls.crt' : nil}"
           ssl_client_cert_key "#{File.exist?('/var/run/ocp-collector/secrets/some-secret/tls.key') ? '/var/run/ocp-collector/secrets/some-secret/tls.key' : nil}"
           <format>
               @type json
           </format>
           <buffer topic>
               @type file
               path '/var/lib/fluentd/kafka_receiver'
               flush_mode interval
               flush_interval 1s
               flush_thread_count 2
               flush_at_shutdown true
               retry_type exponential_backoff
               retry_wait 1s
               retry_max_interval 60s
               retry_forever true
               queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
               total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
               chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
               overflow_action block
           </buffer>
        </match>
        </label>`

		It("should result in a valid kafka label config", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
					Secret: &v1.OutputSecretSpec{
						Name: "some-secret",
					},
				},
			}

			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})
	})

	Context("for kafka output with multiple brokers", func() {
		kafkaConf := `<label @KAFKA_RECEIVER>
	    <match **>
	       @type kafka2
	       brokers broker1-kafka.svc.messaging.cluster.local:9092,broker2-kafka.svc.messaging.cluster.local:9092
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
	           flush_at_shutdown true
             retry_type exponential_backoff
             retry_wait 1s
	           retry_max_interval 60s
	           retry_forever true
             queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
	           total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
	           chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
	           overflow_action block
	       </buffer>
	    </match>
	    </label>`

		It("should produce well formed output label config", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					OutputTypeSpec: v1.OutputTypeSpec{
						Kafka: &v1.Kafka{
							Topic: "topic",
							Brokers: []string{
								"tls://broker1-kafka.svc.messaging.cluster.local:9092",
								"tls://broker2-kafka.svc.messaging.cluster.local:9092",
							},
						},
					},
				},
			}

			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})

		It("should use the default topic if none provided", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					OutputTypeSpec: v1.OutputTypeSpec{
						Kafka: &v1.Kafka{
							Brokers: []string{
								"tls://broker1-kafka.svc.messaging.cluster.local:9092",
								"tls://broker2-kafka.svc.messaging.cluster.local:9092",
							},
						},
					},
				},
			}

			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})
	})

	Context("for broken kafka output configuration", func() {
		It("should return an error if no brokers provided", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					OutputTypeSpec: v1.OutputTypeSpec{
						Kafka: &v1.Kafka{
							Brokers: []string{},
						},
					},
				},
			}
			_, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).Should(HaveOccurred())
		})

		It("should return an error if endpoint not a valid URL", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "not-a-valid-URL",
				},
			}
			_, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).Should(HaveOccurred())
		})
	})
})
