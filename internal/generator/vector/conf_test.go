package vector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

//TODO: Use a detailed CLF spec
var _ = Describe("Testing Complete Config Generation", func() {
	var f = func(testcase generator.ConfGenerateTest) {
		g := generator.MakeGenerator()
		if testcase.Options == nil {
			testcase.Options = generator.Options{}
		}
		e := generator.MergeSections(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, testcase.Options))
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		diff := cmp.Diff(
			strings.Split(strings.TrimSpace(testcase.ExpectedConf), "\n"),
			strings.Split(strings.TrimSpace(conf), "\n"))
		if diff != "" {
			b, _ := json.MarshalIndent(e, "", " ")
			fmt.Printf("elements:\n%s\n", string(b))
			fmt.Println(conf)
			fmt.Printf("diff: %s", diff)
		}
		Expect(diff).To(Equal(""))
	}
	DescribeTable("Generate full vector.toml", f,
		Entry("with complex spec", generator.ConfGenerateTest{
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{
					Fluentd: &logging.FluentdForwarderSpec{
						Buffer: &logging.FluentdBufferSpec{
							ChunkLimitSize: "8m",
							TotalLimitSize: "800000000",
							OverflowAction: "throw_exception",
						},
					},
				},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
[sources.raw_container_logs]
type = "kubernetes_logs"
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"

# Logs from host audit
[sources.host_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/audit/audit.log"]

# Logs from kubernetes audit
[sources.k8s_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/kube-apiserver/audit.log"]

# Logs from openshift audit
[sources.openshift_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]

# Logs from ovn audit
[sources.ovn_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/ovn/acl-audit-log.log"]

[sources.internal_metrics]
type = "internal_metrics"

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = """
  level = "unknown"
  if match!(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn")'){
    level = "warn"
  } else if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'){
    level = "info"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'){
    level = "error"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'){
    level = "debug"
  }
  .level = level
  
  del(.file)
  
  del(.source_type)
  
  del(.stream)
  
  del(.kubernetes.pod_ips)
"""

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = """
  .
"""

[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube")) || (starts_with!(.kubernetes.namespace_name,"openshift")) || (.kubernetes.namespace_name == "default"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube")) || (starts_with!(.kubernetes.namespace_name,"openshift")) || (.kubernetes.namespace_name == "default")'

# Rename log stream to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = """
  .log_type = "application"
"""

# Rename log stream to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = """
  .log_type = "infrastructure"
"""

# Rename log stream to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = """
  .log_type = "audit"
"""

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.mytestapp = '.kubernetes.namespace_name == "test-ns"'

[transforms.pipeline]
type = "remap"
inputs = ["route_application_logs.mytestapp","infrastructure","audit"]
source = """
  .openshift.labels = {"key1":"value1","key2":"value2"}
"""

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

# TLS Config
[sinks.kafka_receiver.tls]
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"
enabled = true

[sinks.prometheus_output]
type = "prometheus_exporter"
inputs = ["internal_metrics"]
address = "0.0.0.0:24231"
default_namespace = "collector"

[sinks.prometheus_output.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
`,
		}),
		Entry("with complex spec for elasticsearch", generator.ConfGenerateTest{
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"es-1", "es-2"},
						Name:       "pipeline",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es-1.svc.messaging.cluster.local:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-2",
						URL:  "https://es-2.svc.messaging.cluster.local:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "es-2",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"es-1": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
				"es-2": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
# Logs from containers (including openshift containers)
[sources.raw_container_logs]
type = "kubernetes_logs"
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"

# Logs from host audit
[sources.host_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/audit/audit.log"]

# Logs from kubernetes audit
[sources.k8s_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/kube-apiserver/audit.log"]

# Logs from openshift audit
[sources.openshift_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]

# Logs from ovn audit
[sources.ovn_audit_logs]
type = "file"
ignore_older_secs = 600
include = ["/var/log/ovn/acl-audit-log.log"]

[sources.internal_metrics]
type = "internal_metrics"

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = """
  level = "unknown"
  if match!(.message,r'(Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn")'){
    level = "warn"
  } else if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'){
    level = "info"
  } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'){
    level = "error"
  } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'){
    level = "debug"
  }
  .level = level
  
  del(.file)
  
  del(.source_type)
  
  del(.stream)
  
  del(.kubernetes.pod_ips)
"""

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = """
  .
"""

[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube")) || (starts_with!(.kubernetes.namespace_name,"openshift")) || (.kubernetes.namespace_name == "default"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube")) || (starts_with!(.kubernetes.namespace_name,"openshift")) || (.kubernetes.namespace_name == "default")'

# Rename log stream to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = """
  .log_type = "application"
"""

# Rename log stream to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = """
  .log_type = "infrastructure"
"""

# Rename log stream to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = """
  .log_type = "audit"
"""

[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = """
  .
"""

# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["pipeline"]
source = """
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
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
        -- delete the "labels" key
        event.log.kubernetes["labels"] = nil
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "%.", "_")
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
"""

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoint = "https://es-1.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"

# TLS Config
[sinks.es_1.tls]
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"

ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"

# Set Elasticsearch index
[transforms.es_2_add_es_index]
type = "remap"
inputs = ["pipeline"]
source = """
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
"""

[transforms.es_2_dedot_and_flatten]
type = "lua"
inputs = ["es_2_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
        -- delete the "labels" key
        event.log.kubernetes["labels"] = nil
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "%.", "_")
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
"""

[sinks.es_2]
type = "elasticsearch"
inputs = ["es_2_dedot_and_flatten"]
endpoint = "https://es-2.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"

# TLS Config
[sinks.es_2.tls]
key_file = "/var/run/ocp-collector/secrets/es-2/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-2/tls.crt"

ca_file = "/var/run/ocp-collector/secrets/es-2/ca-bundle.crt"

[sinks.prometheus_output]
type = "prometheus_exporter"
inputs = ["internal_metrics"]
address = "0.0.0.0:24231"
default_namespace = "collector"

[sinks.prometheus_output.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
`,
		}),
	)
	Describe("test helper functions", func() {
		It("test MakeInputs", func() {
			diff := cmp.Diff(helpers.MakeInputs("a", "b"), "[\"a\",\"b\"]")
			fmt.Println(diff)
			Expect(diff).To(Equal(""))
		})
	})
})
