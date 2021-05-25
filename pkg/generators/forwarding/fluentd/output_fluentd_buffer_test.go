package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"

	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentd config", func() {
	var (
		outputs              []loggingv1.OutputSpec
		defaultForwarderSpec *loggingv1.ForwarderSpec
		customForwarderSpec  *loggingv1.ForwarderSpec
		generator            *ConfigGenerator
	)

	BeforeEach(func() {
		customForwarderSpec = &loggingv1.ForwarderSpec{
			Fluentd: &loggingv1.FluentdForwarderSpec{
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
			var err error
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())

			customForwarderSpec = &loggingv1.ForwarderSpec{
				Fluentd: &loggingv1.FluentdForwarderSpec{
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
		  #flatten labels to prevent field explosion in ES    
		  <filter ** >       
			  @type record_transformer    
			  enable_ruby true    
			  <record>    
				  kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }    
			  </record>        
			  remove_keys $.kubernetes.labels    
		  </filter>		
          <match retry_other_elasticsearch>
            @type copy
            <store>
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
            </store>
          </match>
          <match **>
            @type copy
            <store>
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
            </store>
          </match>
        </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, defaultForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(esConf))
		})
	})

	Context("for output elasticsearch", func() {
		JustBeforeEach(func() {
			var err error
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())

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
		  #flatten labels to prevent field explosion in ES    
		  <filter ** >       
			  @type record_transformer    
			  enable_ruby true    
			  <record>    
				  kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }    
			  </record>        
			  remove_keys $.kubernetes.labels    
		  </filter>		  
          <match retry_other_elasticsearch>
            @type copy
            <store>
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
            </store>
          </match>
          <match **>
            @type copy
            <store>
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
            </store>
          </match>
        </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, defaultForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(esConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
			esConf := `
        <label @OTHER_ELASTICSEARCH>
		  #flatten labels to prevent field explosion in ES    
		  <filter ** >       
			  @type record_transformer    
			  enable_ruby true    
			  <record>    
				  kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }    
			  </record>        
			  remove_keys $.kubernetes.labels    
		  </filter> 		

          <match retry_other_elasticsearch>
            @type copy
            <store>
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
                flush_interval 2s
                flush_thread_count 4
                flush_at_shutdown true
                retry_type periodic
                retry_wait 2s
                retry_max_interval 600s
                retry_forever true
                queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
                total_limit_size 512m
                chunk_limit_size 256m
                overflow_action drop_oldest_chunk
              </buffer>
            </store>
          </match>
          <match **>
            @type copy
            <store>
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
                flush_interval 2s
                flush_thread_count 4
                flush_at_shutdown true
                retry_type periodic
                retry_wait 2s
                retry_max_interval 600s
                retry_forever true
                queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
                total_limit_size 512m
                chunk_limit_size 256m
                overflow_action drop_oldest_chunk
              </buffer>
            </store>
          </match>
        </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, customForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(esConf))
		})
	})

	Context("for output fluentdForward", func() {
		JustBeforeEach(func() {
			var err error
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())

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
            # https://docs.fluentd.org/v1.0/articles/in_forward
            @type forward
            heartbeat_type none
            keepalive true

            <buffer>
            @type file
            path '/var/lib/fluentd/secureforward_receiver'
            queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '1024' }"
            total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
            chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '1m'}"
            flush_mode interval
            flush_interval 5s
            flush_at_shutdown true
            flush_thread_count 2
            retry_type exponential_backoff
            retry_wait 1s
            retry_max_interval 60s
            retry_forever true
            # the systemd journald 0.0.8 input plugin will just throw away records if the buffer
            # queue limit is hit - 'block' will halt further reads and keep retrying to flush the
            # buffer to the remote - default is 'block' because in_tail handles that case
            overflow_action block
            </buffer>

            <server>
            host es.svc.messaging.cluster.local
            port 9654
            </server>
           </match>
         </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, defaultForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(fluentdForwardConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
			fluentdForwardConf := `
      <label @SECUREFORWARD_RECEIVER>
        <match **>
          # https://docs.fluentd.org/v1.0/articles/in_forward
          @type forward
          heartbeat_type none
          keepalive true

          <buffer>
          @type file
          path '/var/lib/fluentd/secureforward_receiver'
          queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '1024' }"
          total_limit_size 512m
          chunk_limit_size 256m
          flush_mode immediate
          flush_interval 2s
          flush_at_shutdown true
          flush_thread_count 4
          retry_type periodic
          retry_wait 2s
          retry_max_interval 600s
          retry_forever true
          # the systemd journald 0.0.8 input plugin will just throw away records if the buffer
          # queue limit is hit - 'block' will halt further reads and keep retrying to flush the
          # buffer to the remote - default is 'block' because in_tail handles that case
          overflow_action drop_oldest_chunk
          </buffer>

          <server>
          host es.svc.messaging.cluster.local
          port 9654
          </server>
        </match>
      </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, customForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(fluentdForwardConf))
		})
	})

	Context("for output syslog", func() {
		JustBeforeEach(func() {
			var err error
			generator, err = NewConfigGenerator(false, false, false)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())

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
            @type copy
            <store>
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
            <buffer >
                @type file
                path '/var/lib/fluentd/syslog_receiver'
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
            </store>
          </match>
        </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, defaultForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(syslogConf))
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
            @type copy
            <store>
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
            <buffer >
                @type file
                path '/var/lib/fluentd/syslog_receiver'
                flush_mode immediate
                flush_interval 2s
                flush_thread_count 4
                flush_at_shutdown true
                retry_type periodic
                retry_wait 2s
                retry_max_interval 600s
                retry_forever true
                queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
                total_limit_size 512m
                chunk_limit_size 256m
                overflow_action drop_oldest_chunk
              </buffer>
            </store>
          </match>
        </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, customForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(syslogConf))
		})
	})

	Context("for output kafka", func() {
		JustBeforeEach(func() {
			var err error
			generator, err = NewConfigGenerator(true, true, true)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())

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

			results, err := generator.generateOutputLabelBlocks(outputs, nil, defaultForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})

		It("should override buffer configuration for given tuning parameters", func() {
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
               flush_mode immediate
               flush_interval 2s
               flush_thread_count 4
               flush_at_shutdown true
               retry_type periodic
               retry_wait 2s
               retry_max_interval 600s
               retry_forever true
               queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
               total_limit_size 512m
               chunk_limit_size 256m
               overflow_action drop_oldest_chunk
           </buffer>
        </match>
        </label>`

			results, err := generator.generateOutputLabelBlocks(outputs, nil, customForwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(kafkaConf))
		})
	})
})
