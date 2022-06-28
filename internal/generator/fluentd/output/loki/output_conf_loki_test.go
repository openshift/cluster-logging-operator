package loki

import (
	"fmt"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"

	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// lokiConfig assembls a loki config block in sections
type lokiConfig struct {
	filter, content, label, buffer string
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
    <buffer>
      %v
    </buffer>
  </match>
</label>
`, c.filter, c.content, c.label, c.buffer)
}

func TestLokiOutput(t *testing.T) {
	var (
		config  lokiConfig
		secrets map[string]*corev1.Secret
		g       generator.Generator
	)

	// testCase runs before/after logic around testFunc
	testCase := func(name string, testFunc func(t *testing.T)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
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
     disable_chunk_backup true
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
						"token": []byte("bearer-token-value"),
					}}}

			var err error
			g = generator.MakeGenerator()
			require.NoError(t, err)
			testFunc(t)
		})
	}

	testCase("insecure configuration", func(t *testing.T) {
		outputs := []v1.OutputSpec{{
			Type: v1.OutputTypeLoki,
			Name: "loki-receiver",
			URL:  "https://logs-us-west1.grafana.net",
		}}
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		require.NoError(t, err)
		config.content = "url https://logs-us-west1.grafana.net"
		require.Equal(t, test.TrimLines(config.String()), test.TrimLines(results), results)
	})

	testCase("secure configuration", func(t *testing.T) {
		outputs := []v1.OutputSpec{{
			Type:   v1.OutputTypeLoki,
			Name:   "loki-receiver",
			URL:    "https://logs-us-west1.grafana.net",
			Secret: &v1.OutputSecretSpec{Name: "a-secret-ref"},
		}}
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		require.NoError(t, err)
		config.content = `url https://logs-us-west1.grafana.net
    username "#{File.read('/var/run/ocp-collector/secrets/a-secret-ref/username') rescue nil}"
    password "#{File.read('/var/run/ocp-collector/secrets/a-secret-ref/password') rescue nil}"
    key '/var/run/ocp-collector/secrets/a-secret-ref/tls.key'
    cert '/var/run/ocp-collector/secrets/a-secret-ref/tls.crt'
    ca_cert '/var/run/ocp-collector/secrets/a-secret-ref/ca-bundle.crt'`
		require.Equal(t, test.TrimLines(config.String()), test.TrimLines(results), results)
	})

	testCase("custom label configuration", func(t *testing.T) {
		outputs := []v1.OutputSpec{{
			Name: "loki-receiver",
			Type: v1.OutputTypeLoki,
			URL:  "https://logs-us-west1.grafana.net",
			OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
				LabelKeys: []string{"kubernetes.labels.app", "kubernetes.container_name"},
			}},
		}}
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		require.NoError(t, err)
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
		require.Equal(t, test.TrimLines(config.String()), test.TrimLines(results), results)
	})

	testCase("applies tenantKey value as Loki tenant", func(t *testing.T) {
		outputs := []v1.OutputSpec{{
			Type: v1.OutputTypeLoki,
			Name: "loki-receiver",
			URL:  "https://logs-us-west1.grafana.net/a-tenant",
			OutputTypeSpec: v1.OutputTypeSpec{
				Loki: &v1.Loki{TenantKey: "foo.bar.baz"},
			},
		}}
		config.content = `url https://logs-us-west1.grafana.net/a-tenant
    tenant ${record.dig("foo","bar","baz")}
`
		es := Conf(nil, secrets["loki-receiver"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		require.NoError(t, err)
		require.Equal(t, test.TrimLines(config.String()), test.TrimLines(results))
	})

	testCase("forward with bearer token", func(t *testing.T) {
		outputs := []v1.OutputSpec{{
			Type:   v1.OutputTypeLoki,
			Name:   "loki-receiver",
			URL:    "https://logs-us-west1.grafana.net",
			Secret: &v1.OutputSecretSpec{Name: "a-secret-ref"},
		}}
		es := Conf(nil, secrets["loki-receiver-token"], outputs[0], generator.NoOptions)
		results, err := g.GenerateConf(es...)
		require.NoError(t, err)
		config.content = `url https://logs-us-west1.grafana.net
    bearer_token_file '/var/run/ocp-collector/secrets/a-secret-ref/token'`
		require.Equal(t, test.TrimLines(config.String()), test.TrimLines(results), results)
	})

}
