package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating fluentd secure forward output store config blocks", func() {

	var (
		err           error
		outputs       []logging.OutputSpec
		forwarderSpec *logging.ForwarderSpec
		generator     *ConfigGenerator
	)
	BeforeEach(func() {
		generator, err = NewConfigGenerator(false, false, true)
		Expect(err).To(BeNil())
	})

	Context("for a secure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type: "fluentdForward",
					Name: "secureforward-receiver",
					URL:  "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-infra-secret",
					},
				},
			}
		})

		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs, forwarderSpec)
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
	     queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '1024' }"
	     total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
	     chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '1m'}"
       flush_mode interval
	     flush_interval 5s       
	     flush_at_shutdown true
	     flush_thread_count 2
       retry_type exponential_backoff
       retry_wait 1s
	     retry_max_interval 300s
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
</label>`))
		})
	})

	Context("for an insecure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type: "fluentdForward",
					Name: "secureforward-receiver",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			}
		})
		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(`<label @SECUREFORWARD_RECEIVER>
			<match **>
				# https://docs.fluentd.org/v1.0/articles/in_forward
			  @type forward

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
				retry_max_interval 300s
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
</label>`))
		})
	})
})
