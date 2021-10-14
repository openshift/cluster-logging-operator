package output_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/fluentdforward"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/syslog"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("outputLabelConf buffer tuning", func() {
	var (
		fluentdSpec *loggingv1.FluentdSpec
	)
	BeforeEach(func() {
		fluentdSpec = &loggingv1.FluentdSpec{
			Tuning: &loggingv1.FluentdTuningSpec{
				Buffer: &loggingv1.FluentdBufferSpec{},
			},
		}
	})
	Context("#RetryTimeout", func() {
		It("should return the default when not configured", func() {
			Expect(output.RetryTimeout(fluentdSpec.Tuning.Buffer)).To(Equal("60m"))
		})
		It("should use the spec'd value when configured", func() {
			fluentdSpec.Tuning.Buffer.RetryTimeout = "72h"
			Expect(output.RetryTimeout(fluentdSpec.Tuning.Buffer)).To(Equal("72h"))
		})
	})

})

var _ = Describe("Generating fluentd config", func() {
	var (
		outputs []loggingv1.OutputSpec
		//defaultForwarderSpec *loggingv1.ForwarderSpec
		customFluentdSpec *loggingv1.FluentdSpec
		g                 generator.Generator
	)

	BeforeEach(func() {
		customFluentdSpec = &loggingv1.FluentdSpec{
			Tuning: &loggingv1.FluentdTuningSpec{
				Buffer: &loggingv1.FluentdBufferSpec{
					ChunkLimitSize:   "256m",
					TotalLimitSize:   "512m",
					OverflowAction:   loggingv1.DropOldestChunkAction,
					FlushThreadCount: 4,
					FlushMode:        loggingv1.FlushModeImmediate,
					FlushInterval:    "2s",
					RetryWait:        "2s",
					RetryMaxInterval: "600s",
					RetryType:        loggingv1.RetryPeriodic,
				},
			},
		}
	})

	Context("for empty forwarder buffer spec", func() {
		JustBeforeEach(func() {
			g = generator.MakeGenerator()

			customFluentdSpec = &loggingv1.FluentdSpec{
				Tuning: &loggingv1.FluentdTuningSpec{
					Buffer: &loggingv1.FluentdBufferSpec{},
				},
			}

			outputs = []loggingv1.OutputSpec{
				{
					Type: loggingv1.OutputTypeElasticsearch,
					Name: "other-elasticsearch",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			}
		})

		It("should provide a default buffer configuration", func() {
			esConf := `
        <label @OTHER_ELASTICSEARCH>
		  #remove structured field if present
		  <filter **>
		    @type record_modifier
		    remove_keys structured
		  </filter>
		  #flatten labels to prevent field explosion in ES
		  <filter **>
			  @type record_transformer
			  enable_ruby true
			  <record>
				  kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }
			  </record>
			  remove_keys $.kubernetes.labels
		  </filter>
          <match retry_other_elasticsearch>
              @type elasticsearch
              @id retry_other_elasticsearch
              host es.svc.messaging.cluster.local
              port 9654
              verify_es_version_at_startup false
              scheme http
              target_index_key viaq_index_name
              id_key viaq_msg_id
              remove_keys viaq_index_name

              type_name _doc
              http_backend typhoeus
              write_operation create
              reload_connections 'true'
              # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
              reload_after '200'
              # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
              sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
              reload_on_failure false
              # 2 ^ 31
              request_timeout 2147483648
              <buffer>
                @type file
                path '/var/lib/fluentd/retry_other_elasticsearch'
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
              </buffer>
          </match>
          <match **>
              @type elasticsearch
              @id other_elasticsearch
              host es.svc.messaging.cluster.local
              port 9654
              verify_es_version_at_startup false
              scheme http
              target_index_key viaq_index_name
              id_key viaq_msg_id
              remove_keys viaq_index_name

              type_name _doc
              retry_tag retry_other_elasticsearch
              http_backend typhoeus
              write_operation create
              reload_connections 'true'
              # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
              reload_after '200'
              # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
              sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
              reload_on_failure false
              # 2 ^ 31
              request_timeout 2147483648
              <buffer>
                @type file
                path '/var/lib/fluentd/other_elasticsearch'
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
              </buffer>
          </match>
        </label>`

			e := elasticsearch.Conf(nil, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(esConf))
		})
	})

	Context("for output elasticsearch", func() {
		JustBeforeEach(func() {
			g = generator.MakeGenerator()

			outputs = []loggingv1.OutputSpec{
				{
					Type: loggingv1.OutputTypeElasticsearch,
					Name: "other-elasticsearch",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			}
		})

		It("should provide a default buffer configuration", func() {
			esConf := `
        <label @OTHER_ELASTICSEARCH>
		  #remove structured field if present
		  <filter **>
		    @type record_modifier
		    remove_keys structured
		  </filter>
		  #flatten labels to prevent field explosion in ES
		  <filter **>
			  @type record_transformer
			  enable_ruby true
			  <record>
				  kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }
			  </record>
			  remove_keys $.kubernetes.labels
		  </filter>
          <match retry_other_elasticsearch>
              @type elasticsearch
              @id retry_other_elasticsearch
              host es.svc.messaging.cluster.local
              port 9654
              verify_es_version_at_startup false
              scheme http
              target_index_key viaq_index_name
              id_key viaq_msg_id
              remove_keys viaq_index_name

              type_name _doc
              http_backend typhoeus
              write_operation create
              reload_connections 'true'
              # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
              reload_after '200'
              # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
              sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
              reload_on_failure false
              # 2 ^ 31
              request_timeout 2147483648
              <buffer>
                @type file
                path '/var/lib/fluentd/retry_other_elasticsearch'
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
              </buffer>
          </match>
          <match **>
              @type elasticsearch
              @id other_elasticsearch
              host es.svc.messaging.cluster.local
              port 9654
              verify_es_version_at_startup false
              scheme http
              target_index_key viaq_index_name
              id_key viaq_msg_id
              remove_keys viaq_index_name

              type_name _doc
              retry_tag retry_other_elasticsearch
              http_backend typhoeus
              write_operation create
              reload_connections 'true'
              # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
              reload_after '200'
              # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
              sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
              reload_on_failure false
              # 2 ^ 31
              request_timeout 2147483648
              <buffer>
                @type file
                path '/var/lib/fluentd/other_elasticsearch'
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
              </buffer>
          </match>
        </label>`

			e := elasticsearch.Conf(nil, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(esConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
			esConf := `
        <label @OTHER_ELASTICSEARCH>
		#remove structured field if present
		<filter **>
		  @type record_modifier
		  remove_keys structured
		</filter>
		  #flatten labels to prevent field explosion in ES
		  <filter **>
			  @type record_transformer
			  enable_ruby true
			  <record>
				  kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }
			  </record>
			  remove_keys $.kubernetes.labels
		  </filter>

          <match retry_other_elasticsearch>
              @type elasticsearch
              @id retry_other_elasticsearch
              host es.svc.messaging.cluster.local
              port 9654
              verify_es_version_at_startup false
              scheme http
              target_index_key viaq_index_name
              id_key viaq_msg_id
              remove_keys viaq_index_name

              type_name _doc
              http_backend typhoeus
              write_operation create
              reload_connections 'true'
              # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
              reload_after '200'
              # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
              sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
              reload_on_failure false
              # 2 ^ 31
              request_timeout 2147483648
              <buffer>
                @type file
                path '/var/lib/fluentd/retry_other_elasticsearch'
                flush_mode immediate
                flush_thread_count 4
                retry_type periodic
                retry_wait 2s
                retry_max_interval 600s
                retry_timeout 60m
                queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
                total_limit_size 512m
                chunk_limit_size 256m
                overflow_action drop_oldest_chunk
              </buffer>
          </match>
          <match **>
              @type elasticsearch
              @id other_elasticsearch
              host es.svc.messaging.cluster.local
              port 9654
              verify_es_version_at_startup false
              scheme http
              target_index_key viaq_index_name
              id_key viaq_msg_id
              remove_keys viaq_index_name

              type_name _doc
              retry_tag retry_other_elasticsearch
              http_backend typhoeus
              write_operation create
              reload_connections 'true'
              # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
              reload_after '200'
              # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
              sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
              reload_on_failure false
              # 2 ^ 31
              request_timeout 2147483648
              <buffer>
                @type file
                path '/var/lib/fluentd/other_elasticsearch'
                flush_mode immediate
                flush_thread_count 4
                retry_type periodic
                retry_wait 2s
                retry_max_interval 600s
                retry_timeout 60m
                queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
                total_limit_size 512m
                chunk_limit_size 256m
                overflow_action drop_oldest_chunk
              </buffer>
          </match>
        </label>`

			e := elasticsearch.Conf(customFluentdSpec.Tuning.Buffer, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(esConf))
		})
	})

	Context("for output fluentdForward", func() {
		JustBeforeEach(func() {
			g = generator.MakeGenerator()

			outputs = []loggingv1.OutputSpec{
				{
					Type: "fluentdForward",
					Name: "secureforward-receiver",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			}
		})

		It("should provide a default buffer configuration", func() {
			fluentdForwardConf := `
        <label @SECUREFORWARD_RECEIVER>
          <match **>
            @type forward
			@id secureforward_receiver
            <server>
            host es.svc.messaging.cluster.local
            port 9654
            </server>
            heartbeat_type none
            keepalive true
            <buffer>
            @type file
            path '/var/lib/fluentd/secureforward_receiver'
            flush_mode interval
            flush_interval 5s
            flush_thread_count 2
            retry_type exponential_backoff
            retry_wait 1s
            retry_max_interval 60s
            retry_timeout 60m
            queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
            total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
            chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
            overflow_action block
            </buffer>

           </match>
         </label>`

			e := fluentdforward.Conf(nil, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(fluentdForwardConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
			fluentdForwardConf := `
      <label @SECUREFORWARD_RECEIVER>
        <match **>
          @type forward
		  @id secureforward_receiver
          <server>
          host es.svc.messaging.cluster.local
          port 9654
          </server>
          heartbeat_type none
          keepalive true

          <buffer>
          @type file
          path '/var/lib/fluentd/secureforward_receiver'
          flush_mode immediate
          flush_thread_count 4
          retry_type periodic
          retry_wait 2s
          retry_max_interval 600s
          retry_timeout 60m
          queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
          total_limit_size 512m
          chunk_limit_size 256m
          overflow_action drop_oldest_chunk
          </buffer>

        </match>
      </label>`

			e := fluentdforward.Conf(customFluentdSpec.Tuning.Buffer, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(fluentdForwardConf))
		})
	})

	Context("for output syslog", func() {
		JustBeforeEach(func() {
			g = generator.MakeGenerator()

			outputs = []loggingv1.OutputSpec{
				{
					Type: loggingv1.OutputTypeSyslog,
					Name: "syslog-receiver",
					URL:  "tcp://sl.svc.messaging.cluster.local:9654",
				},
			}
		})

		It("should provide a default buffer configuration", func() {
			syslogConf := `
		        <label @SYSLOG_RECEIVER>
		          <filter **>
					@type parse_json_field
					json_fields  message
					merge_json_log false
					replace_json_log true
		          </filter>
		          <match **>
		              @type remote_syslog
		              @id syslog_receiver
		              host sl.svc.messaging.cluster.local
		              port 9654
		              rfc rfc5424
		              facility user
		                severity debug
		              protocol tcp
		              packet_size 4096
					  hostname "#{ENV['NODE_NAME']}"
		            timeout 60
		            timeout_exception true
		              keep_alive true
		                keep_alive_idle 75
		                keep_alive_cnt 9
		                keep_alive_intvl 7200
                        <format>
                          @type json
                        </format>
		            <buffer>
		                @type file
		                path '/var/lib/fluentd/syslog_receiver'
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
		              </buffer>
		          </match>
		        </label>`

			e := syslog.Conf(nil, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(syslogConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
			syslogConf := `
		        <label @SYSLOG_RECEIVER>
		          <filter **>
					  @type parse_json_field
					  json_fields  message
					  merge_json_log false
					  replace_json_log true
		          </filter>
		          <match **>
		              @type remote_syslog
		              @id syslog_receiver
		              host sl.svc.messaging.cluster.local
		              port 9654
		              rfc rfc5424
		              facility user
		                severity debug
		              protocol tcp
		              packet_size 4096
					  hostname "#{ENV['NODE_NAME']}"
		            timeout 60
		            timeout_exception true
		              keep_alive true
		                keep_alive_idle 75
		                keep_alive_cnt 9
		                keep_alive_intvl 7200
                        <format>
                          @type json
                        </format>
		            <buffer>
		                @type file
		                path '/var/lib/fluentd/syslog_receiver'
		                flush_mode immediate
		                flush_thread_count 4
		                retry_type periodic
		                retry_wait 2s
		                retry_max_interval 600s
		                retry_timeout 60m
		                queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
		                total_limit_size 512m
		                chunk_limit_size 256m
		                overflow_action drop_oldest_chunk
		              </buffer>
		          </match>
		        </label>`

			e := syslog.Conf(customFluentdSpec.Tuning.Buffer, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(syslogConf))
		})
	})

	Context("for output kafka", func() {
		JustBeforeEach(func() {

			g = generator.MakeGenerator()

			outputs = []loggingv1.OutputSpec{
				{
					Type: loggingv1.OutputTypeKafka,
					Name: "kafka-receiver",
					URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
				},
			}
		})

		It("should provide a default buffer configuration", func() {
			kafkaConf := `<label @KAFKA_RECEIVER>
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
           </buffer>
        </match>
        </label>`

			e := kafka.Conf(nil, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
			kafkaConf := `<label @KAFKA_RECEIVER>
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
               flush_mode immediate
               flush_thread_count 4
               retry_type periodic
               retry_wait 2s
               retry_max_interval 600s
               retry_timeout 60m
               queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
               total_limit_size 512m
               chunk_limit_size 256m
               overflow_action drop_oldest_chunk
           </buffer>
        </match>
        </label>`

			e := kafka.Conf(customFluentdSpec.Tuning.Buffer, nil, outputs[0], nil)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(kafkaConf))
		})
	})
})
