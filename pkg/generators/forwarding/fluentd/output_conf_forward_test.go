package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating fluentd secure forward output store config blocks", func() {

	var (
		err       error
		outputs   []logging.OutputSpec
		generator *ConfigGenerator
	)
	BeforeEach(func() {
		generator, err = NewConfigGenerator(false, false, true)
		Expect(err).To(BeNil())
	})

	Context("for a secure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type: "fluentForward",
					Name: "secureforward-receiver",
					URL:  "es.svc.messaging.cluster.local:9654",
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
			Expect(results[0]).To(EqualTrimLines(`<label @SECUREFORWARD_RECEIVER>
	<match **>
		# https://docs.fluentd.org/v1.0/articles/in_forward
	   @type forward
	   <security>
	     self_hostname "#{ENV['NODE_NAME']}" 
	     shared_key "#{File.open('/var/run/ocp-collector/secrets/my-infra-secret/shared_key') do |f| f.readline end.rstrip}"
	   </security>

	   transport tls
	   tls_verify_hostname false
	   tls_version 'TLSv1_2'
	
	   #tls_client_private_key_path /var/run/ocp-collector/secrets/my-infra-secret/tls.key
	   tls_client_cert_path /var/run/ocp-collector/secrets/my-infra-secret/tls.crt
	   tls_cert_path /var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt

	   <buffer>
	     @type file
	     path '/var/lib/fluentd/secureforward_receiver'
	     queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
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
	</match>
</label>`))
		})
	})

	Context("for an insecure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type: "fluentForward",
					Name: "secureforward-receiver",
					URL:  "es.svc.messaging.cluster.local:9654",
				},
			}
		})
		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(`<label @SECUREFORWARD_RECEIVER>
			<match **>
				# https://docs.fluentd.org/v1.0/articles/in_forward
			  @type forward
	   
			  <buffer>
				@type file
				path '/var/lib/fluentd/secureforward_receiver'
				queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
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
		   </match>
</label>`))
		})
	})
})
