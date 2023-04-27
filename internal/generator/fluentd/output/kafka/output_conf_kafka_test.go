package kafka_test

import (
	"errors"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/kafka"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("Generating external kafka server output store config block", func() {
	var (
		outputs []v1.OutputSpec
		g       generator.Generator
	)
	BeforeEach(func() {
		g = generator.MakeGenerator()
	})

	Context("for a single kafka default output target", func() {
		kafkaConf := `
<label @KAFKA_RECEIVER>
  #dedot namespace_labels and rebuild message field if present
  <filter **>
    @type record_modifier
    <record>
    _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
    </record>
    remove_keys _dummy_, _dummy2_, _dummy3_
  </filter>

  <match **>
    @type kafka2
    @id kafka_receiver
    brokers broker1-kafka.svc.messaging.cluster.local:9092
    default_topic topic
    use_event_time true
    <format>
      @type json
    </format>
    <buffer _topic>
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
`

		It("should result in a valid kafka label config", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
				},
			}

			results, err := g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})

		It("should use the default topic if none provided", func() {
			outputs = []v1.OutputSpec{
				{
					Type: v1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092",
				},
			}

			results, err := g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
	})

	Context("for kafka output with secured communication and authentication", func() {
		kafkaConf := `
<label @KAFKA_RECEIVER>
  #dedot namespace_labels and rebuild message field if present
  <filter **>
    @type record_modifier
    <record>
    _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
    </record>
    remove_keys _dummy_, _dummy2_, _dummy3_
  </filter>

  <match **>
    @type kafka2
    @id kafka_receiver
    brokers broker1-kafka.svc.messaging.cluster.local:9092
    default_topic topic
    use_event_time true
    sasl_over_ssl false
    <format>
      @type json
    </format>
    <buffer _topic>
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
`

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
			secret := &corev1.Secret{
				Data: map[string][]byte{},
			}

			results, err := g.GenerateConf(kafka.Conf(nil, secret, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
		It("should enable Kafka if configured", func() {
			kafkaConf = strings.Replace(kafkaConf, "sasl_over_ssl false", "sasl_over_ssl true", 1)
			secret := &corev1.Secret{
				Data: map[string][]byte{"sasl.enable": nil},
			}
			results, err := g.GenerateConf(kafka.Conf(nil, secret, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
		It("should recognize deprecated SASL key", func() {
			kafkaConf = strings.Replace(kafkaConf, "sasl_over_ssl false", "sasl_over_ssl true", 1)
			secret := &corev1.Secret{
				Data: map[string][]byte{"sasl_over_ssl": nil},
			}
			results, err := g.GenerateConf(kafka.Conf(nil, secret, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
	})

	Context("for kafka output with multiple brokers", func() {
		kafkaConf := `
<label @KAFKA_RECEIVER>
  #dedot namespace_labels and rebuild message field if present
  <filter **>
    @type record_modifier
    <record>
    _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
    </record>
    remove_keys _dummy_, _dummy2_, _dummy3_
  </filter>
  
  <match **>
    @type kafka2
    @id kafka_receiver
    brokers broker1-kafka.svc.messaging.cluster.local:9092,broker2-kafka.svc.messaging.cluster.local:9092
    default_topic topic
    use_event_time true
    <format>
      @type json
    </format>
    <buffer _topic>
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
`

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

			results, err := g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
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

			results, err := g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
	})

	Context("for broken kafka output configuration", func() {
		var err error = nil
		recoverPanic := func(err *error) {
			if e := recover(); e != nil {
				*err = errors.New("conf generation failed")
			}
		}
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
			defer recoverPanic(&err)
			_, err = g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil)...)
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
			defer recoverPanic(&err)
			_, err = g.GenerateConf(kafka.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).Should(HaveOccurred())
		})
	})
})
