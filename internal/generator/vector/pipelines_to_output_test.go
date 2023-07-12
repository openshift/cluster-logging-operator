package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			Pipelines(&clfspec, op),
			Outputs(&clspec, secrets, &clfspec, op),
		)
	}

	DescribeTable("Pipeline(s) to Output(s)", helpers.TestGenerateConfWith(f),
		Entry("Loki Output with no Limit", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Name: logging.OutputTypeLoki,
						Type: logging.OutputTypeLoki,
						URL:  "",
						OutputTypeSpec: logging.OutputTypeSpec{
							Loki: &logging.Loki{},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputTypeLoki},
						Name:       "flow-control",
					},
				},
			},
			ExpectedConf: `
[transforms.flow_control_user_defined]
type = "remap"
inputs = ["application"]
source = '''
	.
'''

[transforms.loki_remap]
type = "remap"
inputs = ["flow_control_user_defined"]
source = '''
	del(.tag)
'''

[transforms.loki_dedot]
type = "lua"
inputs = ["loki_remap"]
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

[sinks.loki]
type = "loki"
inputs = ["loki_dedot"]
endpoint = ""
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.loki.encoding]
codec = "json"

[sinks.loki.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

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
ciphersuites = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"`}),
		Entry("Loki Output with Policy", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Name: logging.OutputTypeLoki,
						Type: logging.OutputTypeLoki,
						URL:  "",
						OutputTypeSpec: logging.OutputTypeSpec{
							Loki: &logging.Loki{},
						},
						Limit: &logging.LimitSpec{
							MaxRecordsPerSecond: 100,
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputTypeLoki},
						Name:       "flow-control",
					},
				},
			},
			ExpectedConf: `
[transforms.flow_control_user_defined]
type = "remap"
inputs = ["application"]
source = '''
	.
'''

[transforms.sink_throttle_loki]
type = "throttle"
inputs = ["flow_control_user_defined"]
window_secs = 1
threshold = 100

[transforms.loki_remap]
type = "remap"
inputs = ["sink_throttle_loki"]
source = '''
	del(.tag)
'''

[transforms.loki_dedot]
type = "lua"
inputs = ["loki_remap"]
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

[sinks.loki]
type = "loki"
inputs = ["loki_dedot"]
endpoint = ""
out_of_order_action = "accept"
healthcheck.enabled = false

[sinks.loki.encoding]
codec = "json"

[sinks.loki.labels]
kubernetes_container_name = "{{kubernetes.container_name}}"
kubernetes_host = "${VECTOR_SELF_NODE_NAME}"
kubernetes_namespace_name = "{{kubernetes.namespace_name}}"
kubernetes_pod_name = "{{kubernetes.pod_name}}"
log_type = "{{log_type}}"

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
ciphersuites = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"`,
		}),
	)
})
