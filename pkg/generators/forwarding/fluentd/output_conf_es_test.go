package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating fluentd config blocks", func() {

	var (
		outputs   []logging.OutputSpec
		generator *ConfigGenerator
		pipeline  logging.PipelineSpec
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
					URL:  "es.svc.messaging.cluster.local:9654",
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
			results, err := generator.generateOutputLabelBlocks(outputs)
			Expect(err).To(BeNil())
			Expect(results[0]).To(EqualTrimLines(`<label @ONCLUSTER_ELASTICSEARCH>
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
			user fluentd
			password changeme

			client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
			client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
			ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
			type_name _doc
			write_operation create
			reload_connections "#{ENV['ES_RELOAD_CONNECTIONS'] || 'true'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#reload-after
			reload_after "#{ENV['ES_RELOAD_AFTER'] || '200'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
			sniffer_class_name "#{ENV['ES_SNIFFER_CLASS_NAME'] || 'Fluent::Plugin::ElasticsearchSimpleSniffer'}"
			reload_on_failure false
			# 2 ^ 31
			request_timeout 2147483648
			<buffer>
				@type file
				path '/var/lib/fluentd/retry_oncluster_elasticsearch'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
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
			user fluentd
			password changeme

			client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
			client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
			ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
			type_name _doc
			retry_tag retry_oncluster_elasticsearch
			write_operation create
			reload_connections "#{ENV['ES_RELOAD_CONNECTIONS'] || 'true'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#reload-after
			reload_after "#{ENV['ES_RELOAD_AFTER'] || '200'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
			sniffer_class_name "#{ENV['ES_SNIFFER_CLASS_NAME'] || 'Fluent::Plugin::ElasticsearchSimpleSniffer'}"
			reload_on_failure false
			# 2 ^ 31
			request_timeout 2147483648
			<buffer>
				@type file
				path '/var/lib/fluentd/oncluster_elasticsearch'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
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
					URL:  "es.svc.messaging.cluster.local:9654",
				},
			}
		})
		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs)
			Expect(err).To(BeNil())
			Expect(results[0]).To(EqualTrimLines(`<label @OTHER_ELASTICSEARCH>
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
			user fluentd
			password changeme

			type_name _doc
			write_operation create
			reload_connections "#{ENV['ES_RELOAD_CONNECTIONS'] || 'true'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#reload-after
			reload_after "#{ENV['ES_RELOAD_AFTER'] || '200'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
			sniffer_class_name "#{ENV['ES_SNIFFER_CLASS_NAME'] || 'Fluent::Plugin::ElasticsearchSimpleSniffer'}"
			reload_on_failure false
			# 2 ^ 31
			request_timeout 2147483648
			<buffer>
				@type file
				path '/var/lib/fluentd/retry_other_elasticsearch'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
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
			user fluentd
			password changeme

			type_name _doc
			retry_tag retry_other_elasticsearch
			write_operation create
			reload_connections "#{ENV['ES_RELOAD_CONNECTIONS'] || 'true'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#reload-after
			reload_after "#{ENV['ES_RELOAD_AFTER'] || '200'}"
			# https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
			sniffer_class_name "#{ENV['ES_SNIFFER_CLASS_NAME'] || 'Fluent::Plugin::ElasticsearchSimpleSniffer'}"
			reload_on_failure false
			# 2 ^ 31
			request_timeout 2147483648
			<buffer>
				@type file
				path '/var/lib/fluentd/other_elasticsearch'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
			</buffer>
		</store>
	</match>
</label>`))
		})
	})
})
