package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	test "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating fluentd config blocks", func() {

	var (
		targets   []logging.PipelineTargetSpec
		generator *ConfigGenerator
	)
	BeforeEach(func() {
		var err error
		generator, err = NewConfigGenerator()
		Expect(err).To(BeNil())
	})

	It("should produce well formed @OUTPUT label match stanza", func() {
		Expect(generator).To(Not(BeNil()))
		results, err := generator.generateSourceMatchBlocks(newSourceTags("logs.app", "**_foo_bar**", "**_xyz_abc**"))
		Expect(err).To(BeNil())
		Expect(len(results) > 0).To(BeTrue())
		test.Expect(results[0]).ToEqual(`<match **_foo_bar** **_xyz_abc**>
		@type relabel
		@label @LOGS_APP
	</match>
	`)
	})

	Context("for a secure endpoint", func() {
		BeforeEach(func() {
			targets = []logging.PipelineTargetSpec{
				{
					Type:     logging.PipelineTargetTypeElasticsearch,
					Endpoint: "es.svc.messaging.cluster.local:9654",
					Certificates: &logging.PipelineTargetCertificatesSpec{
						SecretName: "my-es-secret",
					},
				},
			}
		})

		It("should produce well formed @OUTPUT label match stanza", func() {
			targets = append(targets, logging.PipelineTargetSpec{
				Type:     logging.PipelineTargetTypeElasticsearch,
				Endpoint: "es.svc.second.instance:9654",
				Certificates: &logging.PipelineTargetCertificatesSpec{
					SecretName: "my-other-secret",
				},
			})
			Expect(generator).To(Not(BeNil()))
			results, err := generator.generateLabelCopyBlocks("logs.app", targets)
			Expect(err).To(BeNil())
			Expect(len(results) > 0).To(BeTrue())
			test.Expect(results[0]).ToEqual(`<label @LOGS_APP>
				<match **>
					@type copy
					<store>
						@type relabel
						@label @LOGS_APP_ELASTICSEARCH0
					</store>
					<store>
						@type relabel
						@label @LOGS_APP_ELASTICSEARCH1
					</store>
				</match>
			</label>`)
		})

		It("should produce well formed output label config", func() {
			results, err := generator.generateStoreLabelBlocks("logs.app", targets)
			Expect(err).To(BeNil())
			test.Expect(results[0]).ToEqual(`<label @LOGS_APP_ELASTICSEARCH0>
	<match retry_logs_app>
		@type copy
		<store>
			@type elasticsearch
			@id retry_logs_app_elasticsearch0
			host es.svc.messaging.cluster.local
			port 9654
			scheme https
			ssl_version TLSv1_2
			target_index_key viaq_index_name
			id_key viaq_msg_id
			remove_keys viaq_index_name
			user fluentd
			password changeme

			client_key /var/run/ocp-collector/secrets/my-es-secret/key
			client_cert /var/run/ocp-collector/secrets/my-es-secret/cert
			ca_file /var/run/ocp-collector/secrets/my-es-secret/cacert
			type_name com.redhat.viaq.common
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
				path '/var/lib/fluentd/retry_logs_app_elasticsearch0'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
			</buffer>
		</store>
	</match>
	<match **>
		@type copy
		<store>
			@type elasticsearch
			@id logs_app_elasticsearch0
			host es.svc.messaging.cluster.local
			port 9654
			scheme https
			ssl_version TLSv1_2
			target_index_key viaq_index_name
			id_key viaq_msg_id
			remove_keys viaq_index_name
			user fluentd
			password changeme

			client_key /var/run/ocp-collector/secrets/my-es-secret/key
			client_cert /var/run/ocp-collector/secrets/my-es-secret/cert
			ca_file /var/run/ocp-collector/secrets/my-es-secret/cacert
			type_name com.redhat.viaq.common
			retry_tag retry_logs_app
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
				path '/var/lib/fluentd/logs_app_elasticsearch0'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
			</buffer>
		</store>
	</match>
</label>`)
		})
	})

	Context("for an insecure endpoint", func() {
		BeforeEach(func() {
			targets = []logging.PipelineTargetSpec{
				{
					Type:     logging.PipelineTargetTypeElasticsearch,
					Endpoint: "es.svc.messaging.cluster.local:9654",
				},
			}
		})
		It("should produce well formed output label config", func() {
			results, err := generator.generateStoreLabelBlocks("logs.app", targets)
			Expect(err).To(BeNil())
			test.Expect(results[0]).ToEqual(`<label @LOGS_APP_ELASTICSEARCH0>
	<match retry_logs_app>
		@type copy
		<store>
			@type elasticsearch
			@id retry_logs_app_elasticsearch0
			host es.svc.messaging.cluster.local
			port 9654
			scheme http
			target_index_key viaq_index_name
			id_key viaq_msg_id
			remove_keys viaq_index_name
			user fluentd
			password changeme

			type_name com.redhat.viaq.common
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
				path '/var/lib/fluentd/retry_logs_app_elasticsearch0'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
			</buffer>
		</store>
	</match>
	<match **>
		@type copy
		<store>
			@type elasticsearch
			@id logs_app_elasticsearch0
			host es.svc.messaging.cluster.local
			port 9654
			scheme http
			target_index_key viaq_index_name
			id_key viaq_msg_id
			remove_keys viaq_index_name
			user fluentd
			password changeme

			type_name com.redhat.viaq.common
			retry_tag retry_logs_app
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
				path '/var/lib/fluentd/logs_app_elasticsearch0'
				flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
				flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
				retry_forever true
				queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
			</buffer>
		</store>
	</match>
</label>`)
		})
	})
})
