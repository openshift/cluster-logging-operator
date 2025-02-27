package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("fluentd conf generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Fluentd != nil &&
			clspec.Fluentd.Buffer != nil {
			bufspec = clspec.Fluentd.Buffer
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
			Options: framework.Options{
				framework.IncludeLegacyForwardConfig: "",
			},
			ExpectedConf: `
<label @SECUREFORWARD_RECEIVER>
  #dedot namespace_labels and rebuild message field if present
  <filter **>
    @type record_modifier
    <record>
    _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
    </record>
    remove_keys _dummy_, _dummy2_, _dummy3_
  </filter>

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
    expire_dns_cache 30s
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
      disable_chunk_backup true
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
  #dedot namespace_labels and rebuild message field if present
  <filter **>
    @type record_modifier
    <record>
    _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
    </record>
    remove_keys _dummy_, _dummy2_, _dummy3_
  </filter>

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
    expire_dns_cache 30s
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
      disable_chunk_backup true
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
  #dedot namespace_labels and rebuild message field if present
  <filter **>
    @type record_modifier
    <record>
    _dummy_ ${if m=record.dig("kubernetes","namespace_labels");record["kubernetes"]["namespace_labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy2_ ${if m=record.dig("kubernetes","labels");record["kubernetes"]["labels"]={}.tap{|n|m.each{|k,v|n[k.gsub(/[.\/]/,'_')]=v}};end}
    _dummy3_ ${if m=record.dig("kubernetes","flat_labels");record["kubernetes"]["flat_labels"]=[].tap{|n|m.each_with_index{|s, i|n[i] = s.gsub(/[.\/]/,'_')}};end}
    </record>
    remove_keys _dummy_, _dummy2_, _dummy3_
  </filter>
  
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
    expire_dns_cache 30s
    
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
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
	)
})

func TestFluentdConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluentd Conf Generation")
}
