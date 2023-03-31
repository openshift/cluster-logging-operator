package http

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate fluentd config", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Fluentd != nil &&
			clspec.Fluentd.Buffer != nil {
			bufspec = clspec.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], op)
	}
	DescribeTable("for Http output", helpers.TestGenerateConfWith(f),
		Entry("", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com/logs/app-logs",
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: `
<label @HTTP_RECEIVER>
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
	@type http
	endpoint https://my-logstore.com/logs/app-logs
	http_method post
	content_type "application/x-ndjson"
	<auth>
	  method basic
	  username "#{File.read('/var/run/ocp-collector/secrets/http-receiver/username') rescue nil}"
	  password "#{File.read('/var/run/ocp-collector/secrets/http-receiver/password') rescue nil}"
	</auth>
	<buffer>
	  @type file
	  path '/var/lib/fluentd/http_receiver'
	  flush_mode interval
	  flush_interval 1s
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
		Entry("with Http config", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com/logs/app-logs",
						OutputTypeSpec: v1.OutputTypeSpec{Http: &v1.Http{
							Timeout: "50",
							Headers: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}},
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: `
<label @HTTP_RECEIVER>
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
	@type http
	endpoint https://my-logstore.com/logs/app-logs
	http_method post
	content_type "application/x-ndjson"
	headers {"k1":"v1","k2":"v2"}
	<auth>
	  method basic
	  username "#{File.read('/var/run/ocp-collector/secrets/http-receiver/username') rescue nil}"
	  password "#{File.read('/var/run/ocp-collector/secrets/http-receiver/password') rescue nil}"
	</auth>
	<buffer>
	  @type file
	  path '/var/lib/fluentd/http_receiver'
	  flush_mode interval
	  flush_interval 1s
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
		Entry("with TLS config", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com/logs/app-logs",
						OutputTypeSpec: v1.OutputTypeSpec{Http: &v1.Http{
							Timeout: "50",
							Headers: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}},
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"username":      []byte("username"),
						"password":      []byte("password"),
						"tls.crt":       []byte("-- crt-- "),
						"tls.key":       []byte("-- key-- "),
						"ca-bundle.crt": []byte("-- ca-bundle -- "),
						"passphrase":    []byte("-- passphrase --"),
					},
				},
			},
			ExpectedConf: `
<label @HTTP_RECEIVER>
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
	@type http
	endpoint https://my-logstore.com/logs/app-logs
	http_method post
	content_type "application/x-ndjson"
	headers {"k1":"v1","k2":"v2"}
	<auth>
	  method basic
	  username "#{File.read('/var/run/ocp-collector/secrets/http-receiver/username') rescue nil}"
	  password "#{File.read('/var/run/ocp-collector/secrets/http-receiver/password') rescue nil}"
    </auth>
	tls_private_key_path '/var/run/ocp-collector/secrets/http-receiver/tls.key'
	tls_client_cert_path '/var/run/ocp-collector/secrets/http-receiver/tls.crt'
	tls_ca_cert_path '/var/run/ocp-collector/secrets/http-receiver/ca-bundle.crt'
  tls_client_private_key_passphrase "-- passphrase --" 
	<buffer>
	  @type file
	  path '/var/lib/fluentd/http_receiver'
	  flush_mode interval
	  flush_interval 1s
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

func TestHeaders(t *testing.T) {
	h := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	expected := `{"k1":"v1","k2":"v2"}`
	got := utils.ToHeaderStr(h, "%q:%q")
	if got != expected {
		t.Logf("got: %s, expected: %s", got, expected)
		t.Fail()
	}
}

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
