package vector

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating outputs", func() {
	defaultTLS := "VersionTLS12"
	defaultCiphers := "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Outputs(nil, secrets, &clfspec, op)
	}
	DescribeTable("using #Outputs", helpers.TestGenerateConfWith(f),
		Entry("should honor global minTLSVersion & ciphers with loki as the default logstore regardless of the feature gate setting", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{Name: logging.InputNameApplication, OutputRefs: []string{lokistack.FormatOutputNameFromInput(logging.InputNameApplication)}},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: lokistack.FormatOutputNameFromInput(logging.InputNameApplication),
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				constants.LogCollectorToken: {
					Data: map[string][]byte{
						"token": []byte("token-for-loki"),
					},
				},
			},
			Options: generator.Options{},
			ExpectedConf: `
[transforms.default_loki_apps_remap]
type = "remap"
inputs = ["application"]
source = '''
  del(.tag)
'''

[transforms.default_loki_apps_dedot]
type = "lua"
inputs = ["default_loki_apps_remap"]
version = "2"
hooks.init = "init"
hooks.process = "process"
source = '''
    function init()
        count = 0
    end
    function process(event, emit)
        count = count + 1
        event.log.openshift.sequence = count
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
		dedot(event.log.kubernetes.namespace_labels)
        dedot(event.log.kubernetes.labels)
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "[./]", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
'''

[sinks.default_loki_apps]
type = "loki"
inputs = ["default_loki_apps_dedot"]
endpoint = "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application"
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.default_loki_apps.encoding]
codec = "json"

[sinks.default_loki_apps.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

[sinks.default_loki_apps.tls]
min_tls_version = "` + defaultTLS + `"
ciphersuites = "` + defaultCiphers + `"
ca_file = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"

# Bearer Auth Config
[sinks.default_loki_apps.auth]
strategy = "bearer"
token = "token-for-loki"

[transforms.add_nodename_to_metric]
type = "remap"
inputs = ["internal_metrics"]
source = '''
.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''

[sinks.prometheus_output]
type = "prometheus_exporter"
inputs = ["add_nodename_to_metric"]
address = "[::]:24231"
default_namespace = "collector"

[sinks.prometheus_output.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
min_tls_version = "VersionTLS12"
ciphersuites = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
`,
		}),
	)
})
