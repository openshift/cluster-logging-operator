package loki

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

// lokiConfig assembls a loki config block in sections
type lokiConfig struct {
	filter, content, label, bufferKeys, buffer string
}

func (c *lokiConfig) String() string {
	return fmt.Sprintf(`
<label @LOKI_RECEIVER>
  <filter **>
    @type record_modifier
    <record>
      %v
    </record>
  </filter>
  <match **>
    @type loki
	@id loki_receiver
    line_format json
    %v
    <label>
      %v
    </label>
    <buffer%v>
      %v
    </buffer>
  </match>
</label>
`, c.filter, c.content, c.label, c.bufferKeys, c.buffer)
}

var _ = Describe("Loki output configuration", func() {
	var (
		config  lokiConfig
		secrets map[string]*corev1.Secret
		g       generator.Generator
	)

	BeforeEach(func() {
		config = lokiConfig{ // config with defaults
			filter: `
         _kubernetes_container_name ${record.dig("kubernetes","container_name")}
         _kubernetes_host "#{ENV['NODE_NAME']}"
         _kubernetes_namespace_name ${record.dig("kubernetes","namespace_name")}
         _kubernetes_pod_name ${record.dig("kubernetes","pod_name")}
         _log_type ${record.dig("log_type")}
         _tag ${tag}
`,
			label: `
       kubernetes_container_name _kubernetes_container_name
       kubernetes_host _kubernetes_host
       kubernetes_namespace_name _kubernetes_namespace_name
       kubernetes_pod_name _kubernetes_pod_name
       log_type _log_type
       tag _tag
`,
			buffer: `
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
`,
		}
		secrets = map[string]*corev1.Secret{
			"loki-receiver": {
				Data: map[string][]byte{
					"username":      []byte("my-username"),
					"password":      []byte("my-password"),
					"tls.crt":       []byte("my-tls"),
					"tls.key":       []byte("my-tls-key"),
					"ca-bundle.crt": []byte("my-bundle"),
				}},
			"loki-receiver-token": {
				Data: map[string][]byte{
					"token": []byte("/path/to/token"),
				}}}

		g = generator.MakeGenerator()
	})

	It("generates insecure, default tenant configuration", func() {
		outputs := []loggingv1.OutputSpec{{
			Type:           loggingv1.OutputTypeLoki,
			Name:           "loki-receiver",
			URL:            "https://logs-us-west1.grafana.net",
			OutputTypeSpec: loggingv1.OutputTypeSpec{Loki: &loggingv1.Loki{TenantID: "-"}},
		}}
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		ExpectOK(err)
		config.content = `url https://logs-us-west1.grafana.net`
		Expect(config.String()).To(EqualTrimLines(results))
	})

	It("generates secure, single-tenant configuration", func() {
		outputs := []loggingv1.OutputSpec{{
			Type:           loggingv1.OutputTypeLoki,
			Name:           "loki-receiver",
			URL:            "https://logs-us-west1.grafana.net",
			Secret:         &loggingv1.OutputSecretSpec{Name: "a-secret-ref"},
			OutputTypeSpec: loggingv1.OutputTypeSpec{Loki: &loggingv1.Loki{TenantID: "-"}},
		}}
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		ExpectOK(err)
		config.content = `url https://logs-us-west1.grafana.net
    username "#{File.read('/var/run/ocp-collector/secrets/a-secret-ref/username') rescue nil}"
    password "#{File.read('/var/run/ocp-collector/secrets/a-secret-ref/password') rescue nil}"
    key '/var/run/ocp-collector/secrets/a-secret-ref/tls.key'
    cert '/var/run/ocp-collector/secrets/a-secret-ref/tls.crt'
    ca_cert '/var/run/ocp-collector/secrets/a-secret-ref/ca-bundle.crt'`
		Expect(config.String()).To(EqualTrimLines(results))
	})

	It("generates custom label configuration", func() {
		outputs := []loggingv1.OutputSpec{{
			Name: "loki-receiver",
			Type: loggingv1.OutputTypeLoki,
			URL:  "https://logs-us-west1.grafana.net",
			OutputTypeSpec: loggingv1.OutputTypeSpec{Loki: &loggingv1.Loki{
				LabelKeys: []string{"kubernetes.labels.app", "kubernetes.container_name"},
				TenantID:  "-",
			}},
		}}
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		ExpectOK(err)
		config.content = `url https://logs-us-west1.grafana.net`
		// NOTE: kubernetes.host should be added automatically if not present.
		config.filter = `
      _kubernetes_container_name ${record.dig("kubernetes","container_name")}
      _kubernetes_host "#{ENV['NODE_NAME']}"
      _kubernetes_labels_app ${record.dig("kubernetes","labels","app")}
      _tag ${tag}
`
		config.label = `
      kubernetes_container_name _kubernetes_container_name
      kubernetes_host _kubernetes_host
      kubernetes_labels_app _kubernetes_labels_app
      tag _tag
`
		Expect(config.String()).To(EqualTrimLines(results))
	})

	It("applies tenantID value as Loki tenant", func() {
		outputs := []loggingv1.OutputSpec{{
			Type: loggingv1.OutputTypeLoki,
			Name: "loki-receiver",
			URL:  "https://logs-us-west1.grafana.net",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Loki: &loggingv1.Loki{TenantID: "my-tenant"},
			},
		}}
		config.content = `url https://logs-us-west1.grafana.net
    tenant my-tenant
`
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		ExpectOK(err)
		Expect(config.String()).To(EqualTrimLines(results))
	})

	It("applies tenantKey value as field name", func() {
		outputs := []loggingv1.OutputSpec{{
			Type: loggingv1.OutputTypeLoki,
			Name: "loki-receiver",
			URL:  "https://logs-us-west1.grafana.net",
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Loki: &loggingv1.Loki{TenantKey: "foo.bar.baz"},
			},
		}}
		config.content = `url https://logs-us-west1.grafana.net
    tenant ${$.foo.bar.baz}
`
		config.bufferKeys = ` $.foo.bar.baz`

		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		ExpectOK(err)

		Expect(config.String()).To(EqualTrimLines(results))
	})

	It("forwards with bearer token", func() {
		outputs := []loggingv1.OutputSpec{{
			Type:           loggingv1.OutputTypeLoki,
			Name:           "loki-receiver",
			URL:            "https://logs-us-west1.grafana.net",
			Secret:         &loggingv1.OutputSecretSpec{Name: "a-secret-ref"},
			OutputTypeSpec: loggingv1.OutputTypeSpec{Loki: &loggingv1.Loki{TenantID: "-"}},
		}}
		es := Conf(nil, secrets["loki-receiver-token"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		ExpectOK(err)
		config.content = `url https://logs-us-west1.grafana.net
    bearer_token_file "/path/to/token"`
		Expect(config.String()).To(EqualTrimLines(results))
	})
})
