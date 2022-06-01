package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("fluentd conf generation", func() {
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Forwarder != nil &&
			clspec.Forwarder.Fluentd != nil &&
			clspec.Forwarder.Fluentd.Buffer != nil {
			bufspec = clspec.Forwarder.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], op)
	}
	DescribeTable("for fluentdforward output", helpers.TestGenerateConfWith(f),
		Entry("with tls key,cert,ca-bundle", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeFluentdForward,
						Name: "secureforward-receiver",
						URL:  "https://es.svc.messaging.cluster.local:9654",
						Secret: &logging.OutputSecretSpec{
							Name: "secureforward-receiver-secret",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"secureforward-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			Options: generator.Options{
				generator.IncludeLegacyForwardConfig: "",
			},
			ExpectedConf: `
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
    tls_client_private_key_path '/var/run/ocp-collector/secrets/secureforward-receiver-secret/tls.key'
    tls_client_cert_path '/var/run/ocp-collector/secrets/secureforward-receiver-secret/tls.crt'
    tls_cert_path '/var/run/ocp-collector/secrets/secureforward-receiver-secret/ca-bundle.crt'
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
`,
		}),
		Entry("with shared_key", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeFluentdForward,
						Name: "secureforward-receiver",
						URL:  "https://es.svc.messaging.cluster.local:9654",
						Secret: &logging.OutputSecretSpec{
							Name: "secureforward-receiver-secret",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"secureforward-receiver": {
					Data: map[string][]byte{
						"shared_key": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
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
      shared_key "junk"
    </security>
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
`,
		}),
		Entry("with unsecured, and default port", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeFluentdForward,
						Name: "secureforward-receiver",
						URL:  "http://es.svc.messaging.cluster.local",
						Secret: &logging.OutputSecretSpec{
							Name: "secureforward-receiver-secret",
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
<label @SECUREFORWARD_RECEIVER>
  <match **>
    @type forward
    @id secureforward_receiver
    <server>
      host es.svc.messaging.cluster.local
      port 24224
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
`,
		}),
	)
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
