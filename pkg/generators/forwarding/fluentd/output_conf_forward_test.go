package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	test "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating fluentd secure forward output store config blocks", func() {

	var (
		outputs   []logging.OutputSpec
		generator *ConfigGenerator
	)
	BeforeEach(func() {
		var err error
		generator, err = NewConfigGenerator()
		Expect(err).To(BeNil())
	})

	Context("for a secure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type:     logging.OutputTypeForward,
					Name:     "secureforward-receiver",
					Endpoint: "es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-infra-secret",
					},
				},
			}
		})

		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			test.Expect(results[0]).ToEqual(`<label @SECUREFORWARD_RECEIVER>
	<match **>
	<store>
		# https://docs.fluentd.org/v1.0/articles/in_forward
	   @type forward
	   transport tls
	   <security>
	     self_hostname ocp-clusterlogging-fluentd
	     shared_key secureforward-receiver
	   </security>

	   tls_version #{ENV['FORWARD_TLS_VERSION'] || 'TLSv1_2'}"
	   tls_verify_hostname #{ENV['FORWARD_TLS_VERIFY_HOSTNAME'] || 'false'}"
	   tls_allow_self_signed_cert  #{ENV['FORWARD_TLS_ALLOW_SELF_SIGNED_CERT'] || 'true'}"
	   tls_insecure_mode #{ENV['FORWARD_TLS_INSECURE_MODE'] || 'false'}"
	
	   tls_client_private_key_path /var/run/ocp-collector/secrets/my-infra-secret/tls.key
	   tls_client_cert_path /var/run/ocp-collector/secrets/my-infra-secret/tls.crt
	   tls_cert_path /var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt

	   keepalive #{ENV['FORWARD_KEEPALIVE'] || 'false'}"
	   keepalive_timeout #{ENV['FORWARD_KEEPALIVE_TIMEOUT'] || nil }"

	   send_timeout #{ENV['FORWARD_SEND_TIMEOUT'] || 60 }"
	   connect_timeout #{ENV['FORWARD_CONNECT_TIMEOUT'] || nil }"
	   recover_wait #{ENV['FORWARD_RECOVER_WAIT'] || 10 }"
	   ignore_network_errors_at_startup #{ENV['FORWARD_IGNORE_NETWORK_ERRORS_AT_STARTUP'] || 'false' }"
	   verify_connection_at_startup #{ENV['FORWARD_VERIFY_CONNECTION_AT_STARTUP'] || 'false' }"

	   <buffer>
	     @type file
	     path '/var/lib/fluentd/secureforward_receiver'
	     queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '1024' }"
	     chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '1m' }"
	     flush_interval "#{ENV['FORWARD_FLUSH_INTERVAL'] || '5s'}"
	     flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
	     flush_thread_count "#{ENV['FLUSH_THREAD_COUNT'] || 2}"
	     retry_max_interval "#{ENV['FORWARD_RETRY_WAIT'] || '300'}"
	     retry_forever true
	     # the systemd journald 0.0.8 input plugin will just throw away records if the buffer
	     # queue limit is hit - 'block' will halt further reads and keep retrying to flush the
	     # buffer to the remote - default is 'exception' because in_tail handles that case
	     overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'exception'}"
	   </buffer>

	   <server>
	     host es.svc.messaging.cluster.local
	     port 9654
	   </server>
	 </store>
	</match>
</label>`)
		})
	})

	Context("for an insecure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type:     logging.OutputTypeForward,
					Name:     "secureforward-receiver",
					Endpoint: "es.svc.messaging.cluster.local:9654",
				},
			}
		})
		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			test.Expect(results[0]).ToEqual(`<label @SECUREFORWARD_RECEIVER>
			<match **>
			<store>
				# https://docs.fluentd.org/v1.0/articles/in_forward
			  @type forward
	   
			  keepalive #{ENV['FORWARD_KEEPALIVE'] || 'false'}"
			  keepalive_timeout #{ENV['FORWARD_KEEPALIVE_TIMEOUT'] || nil }"
	   
			  send_timeout #{ENV['FORWARD_SEND_TIMEOUT'] || 60 }"
			  connect_timeout #{ENV['FORWARD_CONNECT_TIMEOUT'] || nil }"
			  recover_wait #{ENV['FORWARD_RECOVER_WAIT'] || 10 }"

			  ignore_network_errors_at_startup #{ENV['FORWARD_IGNORE_NETWORK_ERRORS_AT_STARTUP'] || 'false' }"
			  verify_connection_at_startup #{ENV['FORWARD_VERIFY_CONNECTION_AT_STARTUP'] || 'false' }"
	   
			  <buffer>
				@type file
				path '/var/lib/fluentd/secureforward_receiver'
				queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '1024' }"
				chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '1m' }"
				flush_interval "#{ENV['FORWARD_FLUSH_INTERVAL'] || '5s'}"
				flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
				flush_thread_count "#{ENV['FLUSH_THREAD_COUNT'] || 2}"
				retry_max_interval "#{ENV['FORWARD_RETRY_WAIT'] || '300'}"
				retry_forever true
				# the systemd journald 0.0.8 input plugin will just throw away records if the buffer
				# queue limit is hit - 'block' will halt further reads and keep retrying to flush the
				# buffer to the remote - default is 'exception' because in_tail handles that case
				overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'exception'}"
			  </buffer>
	   
			  <server>
				host es.svc.messaging.cluster.local
				port 9654
			  </server>
			</store>
		   </match>
</label>`)
		})
	})
})
