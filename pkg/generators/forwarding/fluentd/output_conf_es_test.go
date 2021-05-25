package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentd config blocks", func() {

	var (
		outputs       []logging.OutputSpec
		forwarderSpec *logging.ForwarderSpec
		generator     *ConfigGenerator
		pipeline      logging.PipelineSpec
	)
	BeforeEach(func() {
		var err error
		generator, err = NewConfigGenerator(false, false, true)
		Expect(err).To(BeNil())
	})

	Context("for a secure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "oncluster-elasticsearch",
					URL:  "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
			}
			pipeline = logging.PipelineSpec{
				Name:       "my-secure-pipeline",
				InputRefs:  []string{logging.InputNameApplication},
				OutputRefs: []string{"oncluster-elasticsearch"},
			}
		})

		It("should produce well formed @OUTPUT label match stanza", func() {
			pipeline.OutputRefs = append(pipeline.OutputRefs, "other-elasticsearch")
			Expect(generator).To(Not(BeNil()))
			results, err := generator.generatePipelineToOutputLabels([]logging.PipelineSpec{pipeline})
			Expect(err).To(BeNil())
			Expect(len(results) > 0).To(BeTrue())
			Expect(results[0]).To(EqualTrimLines(`<label @MY_SECURE_PIPELINE>
				<match **>
					@type copy
					<store>
						@type relabel
						@label @ONCLUSTER_ELASTICSEARCH
					</store>
					<store>
						@type relabel
						@label @OTHER_ELASTICSEARCH
					</store>
				</match>
			</label>`))
		})

		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(results[0]).To(EqualTrimLines(`<label @ONCLUSTER_ELASTICSEARCH>
	#flatten labels to prevent field explosion in ES    
	<filter ** >       
		@type record_transformer    
		enable_ruby true    
		<record>    
			kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }    
		</record>        
		remove_keys $.kubernetes.labels    
	</filter>
	<match retry_oncluster_elasticsearch>
		@type copy
		<store>
			@type elasticsearch
			@id retry_oncluster_elasticsearch
			host es.svc.messaging.cluster.local
			port 9654
            verify_es_version_at_startup false
			scheme https
			ssl_version TLSv1_2
			target_index_key viaq_index_name
			id_key viaq_msg_id
			remove_keys viaq_index_name

			client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
			client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
			ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
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
				path '/var/lib/fluentd/retry_oncluster_elasticsearch'
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
			@id oncluster_elasticsearch
			host es.svc.messaging.cluster.local
			port 9654
            verify_es_version_at_startup false
			scheme https
			ssl_version TLSv1_2
			target_index_key viaq_index_name
			id_key viaq_msg_id
			remove_keys viaq_index_name

			client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
			client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
			ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
			type_name _doc
			retry_tag retry_oncluster_elasticsearch
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
				path '/var/lib/fluentd/oncluster_elasticsearch'
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
</label>`))
		})
	})

	Context("for an insecure endpoint", func() {
		BeforeEach(func() {
			pipeline.OutputRefs = []string{"other-elasticsearch"}
			outputs = []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "other-elasticsearch",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			}
		})
		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(results[0]).To(EqualTrimLines(`<label @OTHER_ELASTICSEARCH>
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
</label>`))
		})
	})
})
