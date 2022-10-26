package vector

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	"strings"

	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"

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
	var (
		f = func(testcase testhelpers.ConfGenerateTest) {
			g := generator.MakeGenerator()
			if testcase.Options == nil {
				testcase.Options = generator.Options{}
			}
			e := generator.MergeSections(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, constants.OpenshiftNS, testcase.Options))
			conf, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(strings.TrimSpace(testcase.ExpectedConf)).To(matchers.EqualTrimLines(conf))
		}
	)
	DescribeTable("Generate full vector.toml", f,
		Entry("with complex spec", testhelpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{
				Fluentd: &logging.FluentdForwarderSpec{
					Buffer: &logging.FluentdBufferSpec{
						ChunkLimitSize: "8m",
						TotalLimitSize: "800000000",
						OverflowAction: "throw_exception",
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
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"

# Logs from host audit
[sources.raw_host_audit_logs]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from kubernetes audit
[sources.raw_k8s_audit_logs]
type = "file"
include = ["/var/log/kube-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from openshift audit
[sources.raw_openshift_audit_logs]
type = "file"
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from ovn audit
[sources.raw_ovn_audit_logs]
type = "file"
include = ["/var/log/ovn/acl-audit-log.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

[sources.internal_metrics]
type = "internal_metrics"

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
  del(.source_type)
  del(.stream)
  del(.kubernetes.pod_ips)
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".journal.system"
  
  del(.source_type)
  del(._CPU_USAGE_NSEC)
  del(.__REALTIME_TIMESTAMP)
  del(.__MONOTONIC_TIMESTAMP)
  del(._SOURCE_REALTIME_TIMESTAMP)
  del(.PRIORITY)
  del(.JOB_RESULT)
  del(.JOB_TYPE)
  del(.TIMESTAMP_BOOTTIME)
  del(.TIMESTAMP_MONOTONIC)
  
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
  
  .hostname = del(.host)
  
  # systemd’s kernel-specific metadata.
  # .systemd.k = {}
  if exists(.KERNEL_DEVICE) { .systemd.k.KERNEL_DEVICE = del(.KERNEL_DEVICE) }
  if exists(.KERNEL_SUBSYSTEM) { .systemd.k.KERNEL_SUBSYSTEM = del(.KERNEL_SUBSYSTEM) }
  if exists(.UDEV_DEVLINK) { .systemd.k.UDEV_DEVLINK = del(.UDEV_DEVLINK) }
  if exists(.UDEV_DEVNODE) { .systemd.k.UDEV_DEVNODE = del(.UDEV_DEVNODE) }
  if exists(.UDEV_SYSNAME) { .systemd.k.UDEV_SYSNAME = del(.UDEV_SYSNAME) }
  
  # trusted journal fields, fields that are implicitly added by the journal and cannot be altered by client code.
  .systemd.t = {}
  if exists(._AUDIT_LOGINUID) { .systemd.t.AUDIT_LOGINUID = del(._AUDIT_LOGINUID) }
  if exists(._BOOT_ID) { .systemd.t.BOOT_ID = del(._BOOT_ID) }
  if exists(._AUDIT_SESSION) { .systemd.t.AUDIT_SESSION = del(._AUDIT_SESSION) }
  if exists(._CAP_EFFECTIVE) { .systemd.t.CAP_EFFECTIVE = del(._CAP_EFFECTIVE) }
  if exists(._CMDLINE) { .systemd.t.CMDLINE = del(._CMDLINE) }
  if exists(._COMM) { .systemd.t.COMM = del(._COMM) }
  if exists(._EXE) { .systemd.t.EXE = del(._EXE) }
  if exists(._GID) { .systemd.t.GID = del(._GID) }
  if exists(._HOSTNAME) { .systemd.t.HOSTNAME = .hostname }
  if exists(._LINE_BREAK) { .systemd.t.LINE_BREAK = del(._LINE_BREAK) }
  if exists(._MACHINE_ID) { .systemd.t.MACHINE_ID = del(._MACHINE_ID) }
  if exists(._PID) { .systemd.t.PID = del(._PID) }
  if exists(._SELINUX_CONTEXT) { .systemd.t.SELINUX_CONTEXT = del(._SELINUX_CONTEXT) }
  if exists(._SOURCE_REALTIME_TIMESTAMP) { .systemd.t.SOURCE_REALTIME_TIMESTAMP = del(._SOURCE_REALTIME_TIMESTAMP) }
  if exists(._STREAM_ID) { .systemd.t.STREAM_ID = ._STREAM_ID }
  if exists(._SYSTEMD_CGROUP) { .systemd.t.SYSTEMD_CGROUP = del(._SYSTEMD_CGROUP) }
  if exists(._SYSTEMD_INVOCATION_ID) {.systemd.t.SYSTEMD_INVOCATION_ID = ._SYSTEMD_INVOCATION_ID}
  if exists(._SYSTEMD_OWNER_UID) { .systemd.t.SYSTEMD_OWNER_UID = del(._SYSTEMD_OWNER_UID) }
  if exists(._SYSTEMD_SESSION) { .systemd.t.SYSTEMD_SESSION = del(._SYSTEMD_SESSION) }
  if exists(._SYSTEMD_SLICE) { .systemd.t.SYSTEMD_SLICE = del(._SYSTEMD_SLICE) }
  if exists(._SYSTEMD_UNIT) { .systemd.t.SYSTEMD_UNIT = del(._SYSTEMD_UNIT) }
  if exists(._SYSTEMD_USER_UNIT) { .systemd.t.SYSTEMD_USER_UNIT = del(._SYSTEMD_USER_UNIT) }
  if exists(._TRANSPORT) { .systemd.t.TRANSPORT = del(._TRANSPORT) }
  if exists(._UID) { .systemd.t.UID = del(._UID) }
  
  # fields that are directly passed from clients and stored in the journal.
  .systemd.u = {}
  if exists(.CODE_FILE) { .systemd.u.CODE_FILE = del(.CODE_FILE) }
  if exists(.CODE_FUNC) { .systemd.u.CODE_FUNCTION = del(.CODE_FUNC) }
  if exists(.CODE_LINE) { .systemd.u.CODE_LINE = del(.CODE_LINE) }
  if exists(.ERRNO) { .systemd.u.ERRNO = del(.ERRNO) }
  if exists(.MESSAGE_ID) { .systemd.u.MESSAGE_ID = del(.MESSAGE_ID) }
  if exists(.SYSLOG_FACILITY) { .systemd.u.SYSLOG_FACILITY = del(.SYSLOG_FACILITY) }
  if exists(.SYSLOG_IDENTIFIER) { .systemd.u.SYSLOG_IDENTIFIER = del(.SYSLOG_IDENTIFIER) }
  if exists(.SYSLOG_PID) { .systemd.u.SYSLOG_PID = del(.SYSLOG_PID) }
  if exists(.RESULT) { .systemd.u.RESULT = del(.RESULT) }
  if exists(.UNIT) { .systemd.u.UNIT = del(.UNIT) }
  
  .time = format_timestamp!(.timestamp, format: "%FT%T%:z")
  
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.host_audit_logs]
type = "remap"
inputs = ["raw_host_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".linux-audit.log"
  
  match1 = parse_regex(.message, r'type=(?P<type>[^ ]+)') ?? {}
  envelop = {}
  envelop |= {"type": match1.type}
  
  match2, err = parse_regex(.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
  if err == null {
    sp = split(match2.ts_record,":")
    if length(sp) == 2 {
        ts = parse_timestamp(sp[0],"%s.%3f") ?? ""
        envelop |= {"record_id": sp[1]}
        . |= {"audit.linux" : envelop}
        . |= {"@timestamp" : format_timestamp(ts,"%+") ?? ""}
    }
  } else {
    log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
  }
  
  .level = "default"
'''

[transforms.k8s_audit_logs]
type = "remap"
inputs = ["raw_k8s_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".k8s-audit.log"
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
  .k8s_audit_level = .level
  .level = "default"
'''

[transforms.openshift_audit_logs]
type = "remap"
inputs = ["raw_openshift_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".openshift-audit.log"
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
  .openshift_audit_level = .level
  .level = "default"
'''

[transforms.ovn_audit_logs]
type = "remap"
inputs = ["raw_ovn_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".ovn-audit.log"
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
'''

[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.mytestapp = '.kubernetes.namespace_name == "test-ns"'

[transforms.pipeline]
type = "remap"
inputs = ["route_application_logs.mytestapp","infrastructure","audit"]
source = '''
  .openshift.labels = {"key1":"value1","key2":"value2"}
'''

# Kafka config
[sinks.kafka_receiver]
type = "kafka"
inputs = ["pipeline"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "topic"

[sinks.kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"

[sinks.kafka_receiver.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"

[transforms.add_nodename_to_metric]
type = "remap"
inputs = ["internal_metrics"]
source = '''
.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''

[sinks.prometheus_output]
type = "prometheus_exporter"
inputs = ["add_nodename_to_metric"]
address = "0.0.0.0:24231"
default_namespace = "collector"

[sinks.prometheus_output.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
`,
		}),
		Entry("with complex spec for elasticsearch, without version specified", testhelpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{},
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
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"

# Logs from host audit
[sources.raw_host_audit_logs]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from kubernetes audit
[sources.raw_k8s_audit_logs]
type = "file"
include = ["/var/log/kube-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from openshift audit
[sources.raw_openshift_audit_logs]
type = "file"
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from ovn audit
[sources.raw_ovn_audit_logs]
type = "file"
include = ["/var/log/ovn/acl-audit-log.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

[sources.internal_metrics]
type = "internal_metrics"

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
  del(.source_type)
  del(.stream)
  del(.kubernetes.pod_ips)
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  
  .tag = ".journal.system"
  
  del(.source_type)
  del(._CPU_USAGE_NSEC)
  del(.__REALTIME_TIMESTAMP)
  del(.__MONOTONIC_TIMESTAMP)
  del(._SOURCE_REALTIME_TIMESTAMP)
  del(.PRIORITY)
  del(.JOB_RESULT)
  del(.JOB_TYPE)
  del(.TIMESTAMP_BOOTTIME)
  del(.TIMESTAMP_MONOTONIC)
  
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
  
  .hostname = del(.host)
  
  # systemd’s kernel-specific metadata.
  # .systemd.k = {}
  if exists(.KERNEL_DEVICE) { .systemd.k.KERNEL_DEVICE = del(.KERNEL_DEVICE) }
  if exists(.KERNEL_SUBSYSTEM) { .systemd.k.KERNEL_SUBSYSTEM = del(.KERNEL_SUBSYSTEM) }
  if exists(.UDEV_DEVLINK) { .systemd.k.UDEV_DEVLINK = del(.UDEV_DEVLINK) }
  if exists(.UDEV_DEVNODE) { .systemd.k.UDEV_DEVNODE = del(.UDEV_DEVNODE) }
  if exists(.UDEV_SYSNAME) { .systemd.k.UDEV_SYSNAME = del(.UDEV_SYSNAME) }
  
  # trusted journal fields, fields that are implicitly added by the journal and cannot be altered by client code.
  .systemd.t = {}
  if exists(._AUDIT_LOGINUID) { .systemd.t.AUDIT_LOGINUID = del(._AUDIT_LOGINUID) }
  if exists(._BOOT_ID) { .systemd.t.BOOT_ID = del(._BOOT_ID) }
  if exists(._AUDIT_SESSION) { .systemd.t.AUDIT_SESSION = del(._AUDIT_SESSION) }
  if exists(._CAP_EFFECTIVE) { .systemd.t.CAP_EFFECTIVE = del(._CAP_EFFECTIVE) }
  if exists(._CMDLINE) { .systemd.t.CMDLINE = del(._CMDLINE) }
  if exists(._COMM) { .systemd.t.COMM = del(._COMM) }
  if exists(._EXE) { .systemd.t.EXE = del(._EXE) }
  if exists(._GID) { .systemd.t.GID = del(._GID) }
  if exists(._HOSTNAME) { .systemd.t.HOSTNAME = .hostname }
  if exists(._LINE_BREAK) { .systemd.t.LINE_BREAK = del(._LINE_BREAK) }
  if exists(._MACHINE_ID) { .systemd.t.MACHINE_ID = del(._MACHINE_ID) }
  if exists(._PID) { .systemd.t.PID = del(._PID) }
  if exists(._SELINUX_CONTEXT) { .systemd.t.SELINUX_CONTEXT = del(._SELINUX_CONTEXT) }
  if exists(._SOURCE_REALTIME_TIMESTAMP) { .systemd.t.SOURCE_REALTIME_TIMESTAMP = del(._SOURCE_REALTIME_TIMESTAMP) }
  if exists(._STREAM_ID) { .systemd.t.STREAM_ID = ._STREAM_ID }
  if exists(._SYSTEMD_CGROUP) { .systemd.t.SYSTEMD_CGROUP = del(._SYSTEMD_CGROUP) }
  if exists(._SYSTEMD_INVOCATION_ID) {.systemd.t.SYSTEMD_INVOCATION_ID = ._SYSTEMD_INVOCATION_ID}
  if exists(._SYSTEMD_OWNER_UID) { .systemd.t.SYSTEMD_OWNER_UID = del(._SYSTEMD_OWNER_UID) }
  if exists(._SYSTEMD_SESSION) { .systemd.t.SYSTEMD_SESSION = del(._SYSTEMD_SESSION) }
  if exists(._SYSTEMD_SLICE) { .systemd.t.SYSTEMD_SLICE = del(._SYSTEMD_SLICE) }
  if exists(._SYSTEMD_UNIT) { .systemd.t.SYSTEMD_UNIT = del(._SYSTEMD_UNIT) }
  if exists(._SYSTEMD_USER_UNIT) { .systemd.t.SYSTEMD_USER_UNIT = del(._SYSTEMD_USER_UNIT) }
  if exists(._TRANSPORT) { .systemd.t.TRANSPORT = del(._TRANSPORT) }
  if exists(._UID) { .systemd.t.UID = del(._UID) }
  
  # fields that are directly passed from clients and stored in the journal.
  .systemd.u = {}
  if exists(.CODE_FILE) { .systemd.u.CODE_FILE = del(.CODE_FILE) }
  if exists(.CODE_FUNC) { .systemd.u.CODE_FUNCTION = del(.CODE_FUNC) }
  if exists(.CODE_LINE) { .systemd.u.CODE_LINE = del(.CODE_LINE) }
  if exists(.ERRNO) { .systemd.u.ERRNO = del(.ERRNO) }
  if exists(.MESSAGE_ID) { .systemd.u.MESSAGE_ID = del(.MESSAGE_ID) }
  if exists(.SYSLOG_FACILITY) { .systemd.u.SYSLOG_FACILITY = del(.SYSLOG_FACILITY) }
  if exists(.SYSLOG_IDENTIFIER) { .systemd.u.SYSLOG_IDENTIFIER = del(.SYSLOG_IDENTIFIER) }
  if exists(.SYSLOG_PID) { .systemd.u.SYSLOG_PID = del(.SYSLOG_PID) }
  if exists(.RESULT) { .systemd.u.RESULT = del(.RESULT) }
  if exists(.UNIT) { .systemd.u.UNIT = del(.UNIT) }
  
  .time = format_timestamp!(.timestamp, format: "%FT%T%:z")
  
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.host_audit_logs]
type = "remap"
inputs = ["raw_host_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".linux-audit.log"
  
  match1 = parse_regex(.message, r'type=(?P<type>[^ ]+)') ?? {}
  envelop = {}
  envelop |= {"type": match1.type}
  
  match2, err = parse_regex(.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
  if err == null {
    sp = split(match2.ts_record,":")
    if length(sp) == 2 {
        ts = parse_timestamp(sp[0],"%s.%3f") ?? ""
        envelop |= {"record_id": sp[1]}
        . |= {"audit.linux" : envelop}
        . |= {"@timestamp" : format_timestamp(ts,"%+") ?? ""}
    }
  } else {
    log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
  }
  
  .level = "default"
'''

[transforms.k8s_audit_logs]
type = "remap"
inputs = ["raw_k8s_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".k8s-audit.log"
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
  .k8s_audit_level = .level
  .level = "default"
'''

[transforms.openshift_audit_logs]
type = "remap"
inputs = ["raw_openshift_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".openshift-audit.log"
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
  .openshift_audit_level = .level
  .level = "default"
'''

[transforms.ovn_audit_logs]
type = "remap"
inputs = ["raw_ovn_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".ovn-audit.log"
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
'''

[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''

# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["pipeline"]
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
'''

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
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
        flatten_labels(event)
        prune_labels(event)
        emit(event)
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
		local exclusions = {"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component", "app.kubernetes.io/part-of", "app.kubernetes.io/managed-by", "app.kubernetes.io/created-by"}
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

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoint = "https://es-1.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
request.timeout_secs = 2147483648
id_key = "_id"

[sinks.es_1.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"

# Set Elasticsearch index
[transforms.es_2_add_es_index]
type = "remap"
inputs = ["pipeline"]
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
'''

[transforms.es_2_dedot_and_flatten]
type = "lua"
inputs = ["es_2_add_es_index"]
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
        flatten_labels(event)
        prune_labels(event)
        emit(event)
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
		local exclusions = {"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component", "app.kubernetes.io/part-of", "app.kubernetes.io/managed-by", "app.kubernetes.io/created-by"}
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

[sinks.es_2]
type = "elasticsearch"
inputs = ["es_2_dedot_and_flatten"]
endpoint = "https://es-2.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
request.timeout_secs = 2147483648
id_key = "_id"

[sinks.es_2.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-2/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-2/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-2/ca-bundle.crt"

[transforms.add_nodename_to_metric]
type = "remap"
inputs = ["internal_metrics"]
source = '''
.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''

[sinks.prometheus_output]
type = "prometheus_exporter"
inputs = ["add_nodename_to_metric"]
address = "0.0.0.0:24231"
default_namespace = "collector"

[sinks.prometheus_output.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
`,
		}),
		Entry("with complex spec for elasticsearch default v6 & latest version", testhelpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{},
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
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: logging.DefaultESVersion,
							},
						},
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-2",
						URL:  "https://es-2.svc.messaging.cluster.local:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: logging.LatestESVersion,
							},
						},
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
glob_minimum_cooldown_ms = 15000
auto_partial_merge = true
exclude_paths_glob_patterns = ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
pod_annotation_fields.pod_labels = "kubernetes.labels"
pod_annotation_fields.pod_namespace = "kubernetes.namespace_name"
pod_annotation_fields.pod_annotations = "kubernetes.annotations"
pod_annotation_fields.pod_uid = "kubernetes.pod_id"
pod_annotation_fields.pod_node_name = "hostname"

[sources.raw_journal_logs]
type = "journald"
journal_directory = "/var/log/journal"

# Logs from host audit
[sources.raw_host_audit_logs]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from kubernetes audit
[sources.raw_k8s_audit_logs]
type = "file"
include = ["/var/log/kube-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from openshift audit
[sources.raw_openshift_audit_logs]
type = "file"
include = ["/var/log/oauth-apiserver/audit.log","/var/log/openshift-apiserver/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

# Logs from ovn audit
[sources.raw_ovn_audit_logs]
type = "file"
include = ["/var/log/ovn/acl-audit-log.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000

[sources.internal_metrics]
type = "internal_metrics"

[transforms.container_logs]
type = "remap"
inputs = ["raw_container_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
  del(.source_type)
  del(.stream)
  del(.kubernetes.pod_ips)
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.journal_logs]
type = "remap"
inputs = ["raw_journal_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".journal.system"
  
  del(.source_type)
  del(._CPU_USAGE_NSEC)
  del(.__REALTIME_TIMESTAMP)
  del(.__MONOTONIC_TIMESTAMP)
  del(._SOURCE_REALTIME_TIMESTAMP)
  del(.PRIORITY)
  del(.JOB_RESULT)
  del(.JOB_TYPE)
  del(.TIMESTAMP_BOOTTIME)
  del(.TIMESTAMP_MONOTONIC)
  
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
  
  .hostname = del(.host)
  
  # systemd’s kernel-specific metadata.
  # .systemd.k = {}
  if exists(.KERNEL_DEVICE) { .systemd.k.KERNEL_DEVICE = del(.KERNEL_DEVICE) }
  if exists(.KERNEL_SUBSYSTEM) { .systemd.k.KERNEL_SUBSYSTEM = del(.KERNEL_SUBSYSTEM) }
  if exists(.UDEV_DEVLINK) { .systemd.k.UDEV_DEVLINK = del(.UDEV_DEVLINK) }
  if exists(.UDEV_DEVNODE) { .systemd.k.UDEV_DEVNODE = del(.UDEV_DEVNODE) }
  if exists(.UDEV_SYSNAME) { .systemd.k.UDEV_SYSNAME = del(.UDEV_SYSNAME) }
  
  # trusted journal fields, fields that are implicitly added by the journal and cannot be altered by client code.
  .systemd.t = {}
  if exists(._AUDIT_LOGINUID) { .systemd.t.AUDIT_LOGINUID = del(._AUDIT_LOGINUID) }
  if exists(._BOOT_ID) { .systemd.t.BOOT_ID = del(._BOOT_ID) }
  if exists(._AUDIT_SESSION) { .systemd.t.AUDIT_SESSION = del(._AUDIT_SESSION) }
  if exists(._CAP_EFFECTIVE) { .systemd.t.CAP_EFFECTIVE = del(._CAP_EFFECTIVE) }
  if exists(._CMDLINE) { .systemd.t.CMDLINE = del(._CMDLINE) }
  if exists(._COMM) { .systemd.t.COMM = del(._COMM) }
  if exists(._EXE) { .systemd.t.EXE = del(._EXE) }
  if exists(._GID) { .systemd.t.GID = del(._GID) }
  if exists(._HOSTNAME) { .systemd.t.HOSTNAME = .hostname }
  if exists(._LINE_BREAK) { .systemd.t.LINE_BREAK = del(._LINE_BREAK) }
  if exists(._MACHINE_ID) { .systemd.t.MACHINE_ID = del(._MACHINE_ID) }
  if exists(._PID) { .systemd.t.PID = del(._PID) }
  if exists(._SELINUX_CONTEXT) { .systemd.t.SELINUX_CONTEXT = del(._SELINUX_CONTEXT) }
  if exists(._SOURCE_REALTIME_TIMESTAMP) { .systemd.t.SOURCE_REALTIME_TIMESTAMP = del(._SOURCE_REALTIME_TIMESTAMP) }
  if exists(._STREAM_ID) { .systemd.t.STREAM_ID = ._STREAM_ID }
  if exists(._SYSTEMD_CGROUP) { .systemd.t.SYSTEMD_CGROUP = del(._SYSTEMD_CGROUP) }
  if exists(._SYSTEMD_INVOCATION_ID) {.systemd.t.SYSTEMD_INVOCATION_ID = ._SYSTEMD_INVOCATION_ID}
  if exists(._SYSTEMD_OWNER_UID) { .systemd.t.SYSTEMD_OWNER_UID = del(._SYSTEMD_OWNER_UID) }
  if exists(._SYSTEMD_SESSION) { .systemd.t.SYSTEMD_SESSION = del(._SYSTEMD_SESSION) }
  if exists(._SYSTEMD_SLICE) { .systemd.t.SYSTEMD_SLICE = del(._SYSTEMD_SLICE) }
  if exists(._SYSTEMD_UNIT) { .systemd.t.SYSTEMD_UNIT = del(._SYSTEMD_UNIT) }
  if exists(._SYSTEMD_USER_UNIT) { .systemd.t.SYSTEMD_USER_UNIT = del(._SYSTEMD_USER_UNIT) }
  if exists(._TRANSPORT) { .systemd.t.TRANSPORT = del(._TRANSPORT) }
  if exists(._UID) { .systemd.t.UID = del(._UID) }
  
  # fields that are directly passed from clients and stored in the journal.
  .systemd.u = {}
  if exists(.CODE_FILE) { .systemd.u.CODE_FILE = del(.CODE_FILE) }
  if exists(.CODE_FUNC) { .systemd.u.CODE_FUNCTION = del(.CODE_FUNC) }
  if exists(.CODE_LINE) { .systemd.u.CODE_LINE = del(.CODE_LINE) }
  if exists(.ERRNO) { .systemd.u.ERRNO = del(.ERRNO) }
  if exists(.MESSAGE_ID) { .systemd.u.MESSAGE_ID = del(.MESSAGE_ID) }
  if exists(.SYSLOG_FACILITY) { .systemd.u.SYSLOG_FACILITY = del(.SYSLOG_FACILITY) }
  if exists(.SYSLOG_IDENTIFIER) { .systemd.u.SYSLOG_IDENTIFIER = del(.SYSLOG_IDENTIFIER) }
  if exists(.SYSLOG_PID) { .systemd.u.SYSLOG_PID = del(.SYSLOG_PID) }
  if exists(.RESULT) { .systemd.u.RESULT = del(.RESULT) }
  if exists(.UNIT) { .systemd.u.UNIT = del(.UNIT) }
  
  .time = format_timestamp!(.timestamp, format: "%FT%T%:z")
  
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.host_audit_logs]
type = "remap"
inputs = ["raw_host_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".linux-audit.log"
  
  match1 = parse_regex(.message, r'type=(?P<type>[^ ]+)') ?? {}
  envelop = {}
  envelop |= {"type": match1.type}
  
  match2, err = parse_regex(.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
  if err == null {
    sp = split(match2.ts_record,":")
    if length(sp) == 2 {
        ts = parse_timestamp(sp[0],"%s.%3f") ?? ""
        envelop |= {"record_id": sp[1]}
        . |= {"audit.linux" : envelop}
        . |= {"@timestamp" : format_timestamp(ts,"%+") ?? ""}
    }
  } else {
    log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
  }

  .level = "default"
'''

[transforms.k8s_audit_logs]
type = "remap"
inputs = ["raw_k8s_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".k8s-audit.log"
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
  .k8s_audit_level = .level
  .level = "default"
'''

[transforms.openshift_audit_logs]
type = "remap"
inputs = ["raw_openshift_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".openshift-audit.log"
  . = merge(., parse_json!(string!(.message))) ?? .
  del(.message)
  .openshift_audit_level = .level
  .level = "default"
'''

[transforms.ovn_audit_logs]
type = "remap"
inputs = ["raw_ovn_audit_logs"]
source = '''
  .openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"
  .tag = ".ovn-audit.log"
  if !exists(.level) {
    .level = "default"
    if match!(.message, r'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"|<info>') {
      .level = "info"
    } else if match!(.message, r'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"|<error>') {
      .level = "error"
    } else if match!(.message, r'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"|<debug>') {
      .level = "debug"
    } else if match!(.message, r'Notice|NOTICE|^N[0-9]+|level=notice|Value:notice|"level":"notice"|<notice>') {
      .level = "notice"
    } else if match!(.message, r'Alert|ALERT|^A[0-9]+|level=alert|Value:alert|"level":"alert"|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Emergency|EMERGENCY|^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"|<emergency>') {
      .level = "emergency"
    }
  }
'''

[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ts = del(.timestamp); if !exists(."@timestamp") {."@timestamp" = ts}
'''

[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''

# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["pipeline"]
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
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
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
        flatten_labels(event)
        prune_labels(event)
        emit(event)
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
		local exclusions = {"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component", "app.kubernetes.io/part-of", "app.kubernetes.io/managed-by", "app.kubernetes.io/created-by"}
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

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoint = "https://es-1.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
request.timeout_secs = 2147483648
id_key = "_id"

[sinks.es_1.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"

# Set Elasticsearch index
[transforms.es_2_add_es_index]
type = "remap"
inputs = ["pipeline"]
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
  }
'''

[transforms.es_2_dedot_and_flatten]
type = "lua"
inputs = ["es_2_add_es_index"]
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
        flatten_labels(event)
        prune_labels(event)
        emit(event)
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
		local exclusions = {"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component", "app.kubernetes.io/part-of", "app.kubernetes.io/managed-by", "app.kubernetes.io/created-by"}
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

[sinks.es_2]
type = "elasticsearch"
inputs = ["es_2_dedot_and_flatten"]
endpoint = "https://es-2.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
request.timeout_secs = 2147483648
id_key = "_id"
suppress_type_name = true

[sinks.es_2.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-2/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-2/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-2/ca-bundle.crt"

[transforms.add_nodename_to_metric]
type = "remap"
inputs = ["internal_metrics"]
source = '''
.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''

[sinks.prometheus_output]
type = "prometheus_exporter"
inputs = ["add_nodename_to_metric"]
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
