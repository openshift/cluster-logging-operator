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
		e := generator.MergeSections(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, generator.NoOptions))
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
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
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

[sources.raw_journal_logs]
type = "journald"

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
include = ["/var/log/oauth-apiserver.audit.log"]

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  level = "unknown"
  if match(.message,r'(Warning|WARN|W[0-9]+|level=warn|Value:warn|"level":"warn")'){
	level = "warn"
  } else if match(.message, r'Info|INFO|I[0-9]+|level=info|Value:info|"level":"info"'){
	level = "info"
  } else if match(.message, r'Error|ERROR|E[0-9]+|level=error|Value:error|"level":"error"'){
	level = "error"
  } else if match(.message, r'Debug|DEBUG|D[0-9]+|level=debug|Value:debug|"level":"debug"'){
	level = "debug"
  }
  .level = level

  .pipeline_metadata.collector.name = "vector"
  .pipeline_metadata.collector.version = "0.14.1"
  ip4, err = get_env_var("NODE_IPV4")
  .pipeline_metadata.collector.ipaddr4 = ip4
  received, err = format_timestamp(now(),"%+")
  .pipeline_metadata.collector.received_at = received
  .pipeline_metadata.collector.error = err
 '''

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = '''
  level = "unknown"
  if match(.message,r'(Warning|WARN|W[0-9]+|level=warn|Value:warn|"level":"warn")'){
	level = "warn"
  } else if match(.message, r'Info|INFO|I[0-9]+|level=info|Value:info|"level":"info"'){
	level = "info"
  } else if match(.message, r'Error|ERROR|E[0-9]+|level=error|Value:error|"level":"error"'){
	level = "error"
  } else if match(.message, r'Debug|DEBUG|D[0-9]+|level=debug|Value:debug|"level":"debug"'){
	level = "debug"
  }
  .level = level

  .pipeline_metadata.collector.name = "vector"
  .pipeline_metadata.collector.version = "0.14.1"
  ip4, err = get_env_var("NODE_IPV4")
  .pipeline_metadata.collector.ipaddr4 = ip4
  received, err = format_timestamp(now(),"%+")
  .pipeline_metadata.collector.received_at = received
  .pipeline_metadata.collector.error = err
 '''


[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!(starts_with!(.kubernetes.pod_namespace,"kube") || starts_with!(.kubernetes.pod_namespace,"openshift") || .kubernetes.pod_namespace == "default")'
route.infra = 'starts_with!(.kubernetes.pod_namespace,"kube") || starts_with!(.kubernetes.pod_namespace,"openshift") || .kubernetes.pod_namespace == "default"'


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
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs"]
source = """
.log_type = "audit"
"""


[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = """
.
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
`,
		}),
		Entry("with complex spec for elastic-search", generator.ConfGenerateTest{
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

[sources.raw_journal_logs]
type = "journald"

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
include = ["/var/log/oauth-apiserver.audit.log"]

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  level = "unknown"
  if match(.message,r'(Warning|WARN|W[0-9]+|level=warn|Value:warn|"level":"warn")'){
	level = "warn"
  } else if match(.message, r'Info|INFO|I[0-9]+|level=info|Value:info|"level":"info"'){
	level = "info"
  } else if match(.message, r'Error|ERROR|E[0-9]+|level=error|Value:error|"level":"error"'){
	level = "error"
  } else if match(.message, r'Debug|DEBUG|D[0-9]+|level=debug|Value:debug|"level":"debug"'){
	level = "debug"
  }
  .level = level

  .pipeline_metadata.collector.name = "vector"
  .pipeline_metadata.collector.version = "0.14.1"
  ip4, err = get_env_var("NODE_IPV4")
  .pipeline_metadata.collector.ipaddr4 = ip4
  received, err = format_timestamp(now(),"%+")
  .pipeline_metadata.collector.received_at = received
  .pipeline_metadata.collector.error = err
 '''

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = '''
  level = "unknown"
  if match(.message,r'(Warning|WARN|W[0-9]+|level=warn|Value:warn|"level":"warn")'){
	level = "warn"
  } else if match(.message, r'Info|INFO|I[0-9]+|level=info|Value:info|"level":"info"'){
	level = "info"
  } else if match(.message, r'Error|ERROR|E[0-9]+|level=error|Value:error|"level":"error"'){
	level = "error"
  } else if match(.message, r'Debug|DEBUG|D[0-9]+|level=debug|Value:debug|"level":"debug"'){
	level = "debug"
  }
  .level = level

  .pipeline_metadata.collector.name = "vector"
  .pipeline_metadata.collector.version = "0.14.1"
  ip4, err = get_env_var("NODE_IPV4")
  .pipeline_metadata.collector.ipaddr4 = ip4
  received, err = format_timestamp(now(),"%+")
  .pipeline_metadata.collector.received_at = received
  .pipeline_metadata.collector.error = err
 '''


[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!(starts_with!(.kubernetes.pod_namespace,"kube") || starts_with!(.kubernetes.pod_namespace,"openshift") || .kubernetes.pod_namespace == "default")'
route.infra = 'starts_with!(.kubernetes.pod_namespace,"kube") || starts_with!(.kubernetes.pod_namespace,"openshift") || .kubernetes.pod_namespace == "default"'


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
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs"]
source = """
.log_type = "audit"
"""


[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = """
.
"""


# Adding _id field
[transforms.elasticsearch_preprocess]
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
."write-index"=index+"-write"
._id = encode_base64(uuid_v4())
"""

[sinks.es_1]
type = "elasticsearch"
inputs = ["elasticsearch_preprocess"]
endpoint = "https://es-1.svc.messaging.cluster.local:9200"
index = "{{ write-index }}"
request.timeout_secs = 2147483648
bulk_action = "create"
id_key = "_id"
# TLS Config
[sinks.es_1.tls]
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"
[sinks.es_2]
type = "elasticsearch"
inputs = ["elasticsearch_preprocess"]
endpoint = "https://es-2.svc.messaging.cluster.local:9200"
index = "{{ write-index }}"
request.timeout_secs = 2147483648
bulk_action = "create"
id_key = "_id"
# TLS Config
[sinks.es_2.tls]
key_file = "/var/run/ocp-collector/secrets/es-2/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-2/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-2/ca-bundle.crt"
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
