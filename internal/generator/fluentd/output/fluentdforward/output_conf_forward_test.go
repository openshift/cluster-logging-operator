package fluentdforward_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/fluentdforward"
	corev1 "k8s.io/api/core/v1"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentd secure forward output store config blocks", func() {

	var (
		outputs []logging.OutputSpec
		g       generator.Generator
	)
	BeforeEach(func() {
		g = generator.MakeGenerator()
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
				outputs[0].Name: {
					Data: map[string][]byte{
						"shared_key":    []byte("my-key"),
						"tls.crt":       []byte("my-tls"),
						"tls.key":       []byte("my-tls-key"),
						"ca-bundle.crt": []byte("my-bundle"),
						"passphrase":    []byte("my-tls-passphrase"),
					},
				},
			}
		})

		It("should skip missing secrets in the config", func() {
			data := secrets[outputs[0].Name].Data
			delete(data, "shared_key")
			delete(data, "tls.key")
			delete(data, "passphrase")
			secret := secrets[outputs[0].Name]
			results, err := g.GenerateConf(fluentdforward.Conf(nil, secret, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    keepalive_timeout 30s
    transport tls
    tls_verify_hostname false
    tls_version 'TLSv1_2'
    tls_cert_path '/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt'
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
</label>
`))
		})

		It("should use insecure mode if no secret", func() {
			outputs[0].Secret = nil
			results, err := g.GenerateConf(fluentdforward.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    keepalive_timeout 30s
    transport tls
    tls_verify_hostname false
    tls_version 'TLSv1_2'
    tls_insecure_mode true
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
</label>
`))
		})

		It("should produce well formed output label config", func() {
			secret := secrets[outputs[0].Name]
			results, err := g.GenerateConf(fluentdforward.Conf(nil, secret, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    keepalive_timeout 30s
    transport tls
    tls_verify_hostname false
    tls_version 'TLSv1_2'
    <security>
      self_hostname "#{ENV['NODE_NAME']}"
      shared_key "my-key"
    </security>
    tls_client_private_key_path '/var/run/ocp-collector/secrets/my-infra-secret/tls.key'
    tls_client_cert_path '/var/run/ocp-collector/secrets/my-infra-secret/tls.crt'
    tls_cert_path '/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt'
    tls_client_private_key_passphrase "#{File.exists?('/var/run/ocp-collector/secrets/my-infra-secret/passphrase') ? open('/var/run/ocp-collector/secrets/my-infra-secret/passphrase','r') do |f|f.read end : ''}"
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
</label>
`))
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
			results, err := g.GenerateConf(fluentdforward.Conf(nil, nil, outputs[0], nil)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
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
    keepalive_timeout 30s
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
</label>
`))
		})
	})
})
