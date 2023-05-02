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
[transforms.flow-control]
type = "remap"
inputs = ["application"]
source = '''
  .
'''

[transforms.loki_remap]
type = "remap"
inputs = ["flow-control"]
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
		Entry("Loki Output with Drop Policy", helpers.ConfGenerateTest{
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
							Policy:              logging.DropPolicy,
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
[transforms.flow-control]
type = "remap"
inputs = ["application"]
source = '''
	.
'''


[transforms.sink_throttle_loki]
type = "throttle"
inputs = ["flow-control"]
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
		Entry("Loki with Ignore Policy", helpers.ConfGenerateTest{
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
							Policy:              logging.DropPolicy,
							MaxRecordsPerSecond: 0,
						},
					},
					{
						Name: logging.OutputTypeCloudwatch,
						Type: logging.OutputTypeCloudwatch,
						URL:  "",
						OutputTypeSpec: logging.OutputTypeSpec{
							Cloudwatch: &logging.Cloudwatch{},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputTypeLoki, logging.OutputTypeCloudwatch},
						Name:       "flow-control",
					},
				},
			},
			ExpectedConf: `
[transforms.flow-control]
type = "remap"
inputs = ["application"]
source = '''
	.
'''

# Cloudwatch Group and Stream Names
[transforms.cloudwatch_normalize_group_and_streams]
type = "remap"
inputs = ["flow-control"]
source = '''
	.group_name = "default"
	.stream_name = "default"
	
	if (.file != null) {
	.file = "kubernetes" + replace!(.file, "/", ".")
	.stream_name = del(.file)
	}
	
	if ( .log_type == "application" ) {
	.group_name = ( "" + .log_type ) ?? "application"
	}
	if ( .log_type == "audit" ) {
	.group_name = "audit"
	.stream_name = ( "${VECTOR_SELF_NODE_NAME}" + .tag ) ?? .stream_name
	}
	if ( .log_type == "infrastructure" ) {
	.group_name = "infrastructure"
	.stream_name = ( .hostname + "." + .stream_name ) ?? .stream_name
	}
	if ( .tag == ".journal.system" ) {
	.stream_name =  ( .hostname + .tag ) ?? .stream_name
	}
	del(.tag)
	del(.source_type)
'''

[transforms.cloudwatch_dedot]
type = "lua"
inputs = ["cloudwatch_normalize_group_and_streams"]
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

# Cloudwatch Logs
[sinks.cloudwatch]
type = "aws_cloudwatch_logs"
inputs = ["cloudwatch_dedot"]
region = ""
compression = "none"
group_name = "{{ group_name }}"
stream_name = "{{ stream_name }}"
auth.access_key_id = ""
auth.secret_access_key = ""
encoding.codec = "json"
request.concurrency = 2
healthcheck.enabled = false

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
		Entry("Loki with Ignore and ES with Drop Policy", helpers.ConfGenerateTest{
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
							Policy:              logging.DropPolicy,
							MaxRecordsPerSecond: 0,
						},
					},
					{
						Name: logging.OutputTypeElasticsearch,
						Type: logging.OutputTypeElasticsearch,
						URL:  "",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{},
						},
						Limit: &logging.LimitSpec{
							Policy:              logging.DropPolicy,
							MaxRecordsPerSecond: 100,
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication},
						OutputRefs: []string{logging.OutputTypeLoki, logging.OutputTypeElasticsearch},
						Name:       "flow-control",
					},
				},
			},
			ExpectedConf: `
[transforms.flow-control]
type = "remap"
inputs = ["application"]
source = '''
	.
'''

[transforms.sink_throttle_elasticsearch]
type = "throttle"
inputs = ["flow-control"]
window_secs = 1
threshold = 100

# Set Elasticsearch index
[transforms.elasticsearch_add_es_index]
type = "remap"
inputs = ["sink_throttle_elasticsearch"]
source = '''
	index = "default"
	if (.log_type == "application"){
	index = "app"
	}
	if (.log_type == "infrastructure"){
	index = "infra"
	}
	if (.log_type == "audit"){
	index = "audit"
	}
	.write_index = index + "-write"
	._id = encode_base64(uuid_v4())
	del(.file)
	del(.tag)
	del(.source_type)
	if .structured != null && .write_index == "app-write" {
		.message = encode_json(.structured)
		del(.structured)
	}
'''

[transforms.elasticsearch_dedot_and_flatten]
type = "lua"
inputs = ["elasticsearch_add_es_index"]
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
		flatten_labels(event)
		prune_labels(event)
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

	function flatten_labels(event)
		-- create "flat_labels" key
		event.log.kubernetes.flat_labels = {}
		i = 1
		-- flatten the labels
		for k,v in pairs(event.log.kubernetes.labels) do
			event.log.kubernetes.flat_labels[i] = k.."="..v
			i=i+1
		end
	end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
	end
'''

[sinks.elasticsearch]
type = "elasticsearch"
inputs = ["elasticsearch_dedot_and_flatten"]
endpoint = ""
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
request.timeout_secs = 2147483648
id_key = "_id"

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
