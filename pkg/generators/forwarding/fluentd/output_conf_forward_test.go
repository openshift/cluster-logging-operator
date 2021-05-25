package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
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
		var secrets map[string]*corev1.Secret

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
			secrets = map[string]*corev1.Secret{
				outputs[0].Secret.Name: {
					Data: map[string][]byte{
						"shared_key":    []byte("my-key"),
						"tls.crt":       []byte("my-tls"),
						"tls.key":       []byte("my-tls-key"),
						"ca-bundle.crt": []byte("my-bundle"),
					},
				},
			}
		})

		It("should skip missing secrets in the config", func() {
			data := secrets["my-infra-secret"].Data
			delete(data, "shared_key")
			delete(data, "tls.key")
			results, err := generator.generateOutputLabelBlocks(outputs, secrets, nil)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(`    <label @SECUREFORWARD_RECEIVER>
      <match **>
        # https://docs.fluentd.org/v1.0/articles/in_forward
        @type forward
        heartbeat_type none
        keepalive true
        transport tls
        tls_verify_hostname false
        tls_version 'TLSv1_2'
        tls_client_cert_path "/var/run/ocp-collector/secrets/my-infra-secret/tls.crt"
        tls_cert_path "/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt"

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
    </label>
`))
		})

		It("should use insecure mode if no secret", func() {
			outputs[0].Secret = nil
			results, err := generator.generateOutputLabelBlocks(outputs, nil, nil)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(`    <label @SECUREFORWARD_RECEIVER>
      <match **>
        # https://docs.fluentd.org/v1.0/articles/in_forward
        @type forward
        heartbeat_type none
        keepalive true
        transport tls
        tls_verify_hostname false
        tls_version 'TLSv1_2'
        tls_insecure_mode true

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
    </label>
`))
		})

		It("should produce well formed output label config", func() {
			results, err := generator.generateOutputLabelBlocks(outputs, secrets, nil)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(`<label @SECUREFORWARD_RECEIVER>
  <match **>
     # https://docs.fluentd.org/v1.0/articles/in_forward
     @type forward
     heartbeat_type none
     keepalive true
     <security>
       self_hostname "#{ENV['NODE_NAME']}"
       shared_key "my-key"
     </security>

     transport tls
     tls_verify_hostname false
     tls_version 'TLSv1_2'

     tls_client_private_key_path "/var/run/ocp-collector/secrets/my-infra-secret/tls.key"
     tls_client_cert_path "/var/run/ocp-collector/secrets/my-infra-secret/tls.crt"
     tls_cert_path "/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt"

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
			results, err := generator.generateOutputLabelBlocks(outputs, nil, nil)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(`<label @SECUREFORWARD_RECEIVER>
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
</label>`))
		})
	})
})
