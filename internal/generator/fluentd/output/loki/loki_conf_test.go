package loki

import (
	"sort"
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("outputLabelConf", func() {
	var (
		loki *logging.Loki
	)
	BeforeEach(func() {
		loki = &logging.Loki{}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(defaultLabelKeys, requiredLabelKeys...)
				sort.Strings(exp)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})
		Context("are spec'd", func() {

			It("should use the ones provided and add the required ones", func() {
				loki.LabelKeys = []string{"foo"}
				exp := append(loki.LabelKeys, requiredLabelKeys...)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})

	})
	Context("#setTLSProfileFromOptions", func() {
		var (
			op           generator.Options
			lokiTemplate Loki
		)
		BeforeEach(func() {
			lokiTemplate = Loki{}
			op = generator.Options{}
		})
		It("should set the ciphers", func() {
			ciphers := "abc,123"
			op[generator.Ciphers] = ciphers
			lokiTemplate.setTLSProfileFromOptions(op)
			Expect(lokiTemplate.CipherSuites).To(Equal(ciphers))
		})
		DescribeTable("should convert the TLS min_version", func(version, exp string) {
			op[generator.MinTLSVersion] = version
			lokiTemplate.setTLSProfileFromOptions(op)
			Expect(lokiTemplate.TLSMinVersion).To(Equal(exp))
		},
			Entry(" for VersionTLS10 it should upgrade to 1.1", "VersionTLS10", "TLS1_1"),
			Entry(" for VersionTLS11", "VersionTLS11", "TLS1_1"),
			Entry(" for VersionTLS12", "VersionTLS12", "TLS1_2"),
			Entry(" for VersionTLS13", "VersionTLS13", "TLS1_3"),
		)
	})
})

var _ = Describe("[internal][generator][fluentd][output][loki] #Conf", func() {
	defaultTLS := "VersionTLS12"
	defaultCiphers := "TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Fluentd != nil &&
			clspec.Fluentd.Buffer != nil {
			bufspec = clspec.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], op)
	}
	DescribeTable("for Loki output", helpers.TestGenerateConfWith(f),
		Entry("with TLS Profile", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Options: generator.Options{
				generator.MinTLSVersion: defaultTLS,
				generator.Ciphers:       defaultCiphers,
			},
			ExpectedConf: `
<label @LOKI_RECEIVER>
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
  
  <filter **>
    @type record_modifier
    <record>
      _kubernetes_container_name ${record.dig("kubernetes","container_name")}
      _kubernetes_host "#{ENV['NODE_NAME']}"
      _kubernetes_namespace_name ${record.dig("kubernetes","namespace_name")}
      _kubernetes_pod_name ${record.dig("kubernetes","pod_name")}
      _log_type ${record.dig("log_type")}
    </record>
  </filter>
  
  <match **>
    @type loki
    @id loki_receiver
    line_format json
    url https://logs-us-west1.grafana.net
    min_version TLS1_2
	ciphers TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384
    <label>
      kubernetes_container_name _kubernetes_container_name
      kubernetes_host _kubernetes_host
      kubernetes_namespace_name _kubernetes_namespace_name
      kubernetes_pod_name _kubernetes_pod_name
      log_type _log_type
    </label>
    <buffer>
      @type file
      path '/var/lib/fluentd/loki_receiver'
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
		Entry("with default labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("junk"),
						"password": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
<label @LOKI_RECEIVER>
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
  
  <filter **>
    @type record_modifier
    <record>
      _kubernetes_container_name ${record.dig("kubernetes","container_name")}
      _kubernetes_host "#{ENV['NODE_NAME']}"
      _kubernetes_namespace_name ${record.dig("kubernetes","namespace_name")}
      _kubernetes_pod_name ${record.dig("kubernetes","pod_name")}
      _log_type ${record.dig("log_type")}
    </record>
  </filter>
  
  <match **>
    @type loki
    @id loki_receiver
    line_format json
    url https://logs-us-west1.grafana.net
    username "#{File.read('/var/run/ocp-collector/secrets/es-1/username') rescue nil}"
    password "#{File.read('/var/run/ocp-collector/secrets/es-1/password') rescue nil}"
    <label>
      kubernetes_container_name _kubernetes_container_name
      kubernetes_host _kubernetes_host
      kubernetes_namespace_name _kubernetes_namespace_name
      kubernetes_pod_name _kubernetes_pod_name
      log_type _log_type
    </label>
    <buffer>
      @type file
      path '/var/lib/fluentd/loki_receiver'
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
		Entry("with custom labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
						OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
							LabelKeys: []string{"kubernetes.labels.app", "kubernetes.container_name"},
						}},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("junk"),
						"password": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
<label @LOKI_RECEIVER>
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
  
  <filter **>
    @type record_modifier
    <record>
      _kubernetes_container_name ${record.dig("kubernetes","container_name")}
      _kubernetes_host "#{ENV['NODE_NAME']}"
      _kubernetes_labels_app ${record.dig("kubernetes","labels","app")}
    </record>
  </filter>
  
  <match **>
    @type loki
    @id loki_receiver
    line_format json
    url https://logs-us-west1.grafana.net
    username "#{File.read('/var/run/ocp-collector/secrets/es-1/username') rescue nil}"
    password "#{File.read('/var/run/ocp-collector/secrets/es-1/password') rescue nil}"
    <label>
      kubernetes_container_name _kubernetes_container_name
      kubernetes_host _kubernetes_host
      kubernetes_labels_app _kubernetes_labels_app
    </label>
    <buffer>
      @type file
      path '/var/lib/fluentd/loki_receiver'
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

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
