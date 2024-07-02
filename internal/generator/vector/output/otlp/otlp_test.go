package otlp

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generate vector config", func() {
	DescribeTable("for OTLP output", func(output obs.OutputSpec, secret helpers.Secrets, op framework.Options, exp string) {
		conf := New(helpers.MakeOutputID(output.Name), output, []string{"pipeline_my_pipeline_viaq_0"}, secret, nil, op) //, includeNS, excludes)
		Expect(exp).To(EqualConfigFrom(conf))
	},
		Entry("with basic spec",
			obs.OutputSpec{
				Type: logging.OutputTypeOtlp,
				Name: "otel-collector",
				OTLP: &obs.OTLP{
					URL: "http://localhost:4318/v1/logs",
				},
			},
			nil,
			framework.NoOptions,
			`
# Route container, journal, and audit logs separately
[transforms.output_otel_collector_reroute]
type = "route"
inputs = ["pipeline_my_pipeline_viaq_0"]
route.container = '.log_source == "container"'
route.journal = '.log_source == "node"'
route.linux = '.log_source == "auditd"'
route.kube = '.log_source == "kubeAPI"'
route.openshift = '.log_source == "openshiftAPI"'
route.ovn = '.log_source == "ovn"'

# Normalize container log records to OTLP semantic conventions
[transforms.output_otel_collector_otlp_container]
type = "remap"
inputs = ["output_otel_collector_reroute.container"]
source = '''
  # Create base resource attributes
  resource.attributes = []
  resource.attributes = append( resource.attributes, 
      [{"key": "node.name", "value": {"stringValue": .hostname}},
      {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
  )
  # Append container resource attributes
  resource.attributes = append( resource.attributes,
      [{"key": "k8s.pod.name", "value": {"stringValue": get!(.,["kubernetes","pod_name"])}},
  	{"key": "k8s.pod.uid", "value": {"stringValue": get!(.,["kubernetes","pod_id"])}},
  	{"key": "k8s.container.name", "value": {"stringValue": get!(.,["kubernetes","container_name"])}},
  	{"key": "k8s.container.id", "value": {"stringValue": get!(.,["kubernetes","container_id"])}},
  	{"key": "k8s.namespace.name", "value": {"stringValue": get!(.,["kubernetes","namespace_name"])}}]
  )
  # Append kube pod labels
  if exists(.kubernetes.labels) {
      for_each(object!(.kubernetes.labels)) -> |key,value| {  
  	    resource.attributes = append(resource.attributes,
  		    [{"key": "k8s.pod.label." + key, "value": {"stringValue": value}}]
  	    )
      }
  }
  # Create logRecord object
  r = {}
  r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
  r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
  # Convert syslog severity keyword to number, default to 9 (unknown)
  r.severityNumber = to_syslog_severity(.level) ?? 9
  r.body = {"stringValue": string!(.message)}
  r.attributes = []
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "openshift.log.type", "value": {"stringValue": .log_type}},
  	{"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
  )
  # Openshift and kubernetes objects for grouping containers (dropped before sending)
  o = {
      "log_type": .log_type,
      "log_source": .log_source,
      "cluster_id": get!(.,["openshift","cluster_id"])
  }
  .kubernetes = {
      "namespace_name": .kubernetes.namespace_name,
      "pod_name": .kubernetes.pod_name,
      "container_name": .kubernetes.container_name
  }
  . = {
    "openshift": o,
    "kubernetes": .kubernetes,
    "resource": resource,
    "logRecords": r
  }
'''

# Merge container logs and group by namespace, pod and container
[transforms.output_otel_collector_group_by_container]
type = "reduce"
inputs = ["output_otel_collector_otlp_container"]
expire_after_ms = 10000
max_events = 3
group_by = [".openshift.cluster_id",".kubernetes.namespace_name",".kubernetes.pod_name",".kubernetes.container_name"]
merge_strategies.resource = "retain"
merge_strategies.logRecords = "array"

# Normalize node log events to OTLP semantic conventions
[transforms.output_otel_collector_otlp_node]
type = "remap"
inputs = ["output_otel_collector_reroute.journal"]
source = '''
  # Create base resource attributes
  resource.attributes = []
  resource.attributes = append( resource.attributes, 
      [{"key": "node.name", "value": {"stringValue": .hostname}},
      {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
  )
  # Create logRecord object
  r = {}
  r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
  r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
  # Convert syslog severity keyword to number, default to 9 (unknown)
  r.severityNumber = to_syslog_severity(.level) ?? 9
  r.body = {"stringValue": string!(.message)}
  r.attributes = []
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "openshift.log.type", "value": {"stringValue": .log_type}},
  	{"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
  )
  # Append log attributes for node logs
  logAttribute = [
    "systemd.t.BOOT_ID",
    "systemd.t.COMM",
    "systemd.t.CAP_EFFECTIVE",
    "systemd.t.CMDLINE",
    "systemd.t.COMM",
    "systemd.t.EXE",
    "systemd.t.GID",
    "systemd.t.MACHINE_ID",
    "systemd.t.PID",
    "systemd.t.SELINUX_CONTEXT",
    "systemd.t.STREAM_ID",
    "systemd.t.SYSTEMD_CGROUP",
    "systemd.t.SYSTEMD_INVOCATION_ID",
    "systemd.t.SYSTEMD_SLICE",
    "systemd.t.SYSTEMD_UNIT",
    "systemd.t.TRANSPORT",
    "systemd.t.UID",
    "systemd.u.SYSLOG_FACILITY",
    "systemd.u.SYSLOG_IDENTIFIER",
  ]
  replacements = {
    "SYSTEMD.CGROUP": "system.cgroup",
    "SYSTEMD.INVOCATION.ID": "system.invocation.id",
    "SYSTEMD.SLICE": "system.slice",
    "SYSTEMD.UNIT": "system.unit",
    "SYSLOG.FACILITY": "syslog.facility",
    "SYSLOG.IDENTIFIER": "syslog.identifier",
    "PID": "syslog.procid",
    "STREAM_ID": "syslog.procid"
  }
  for_each(logAttribute) -> |_,sub_key| {
    path = split(sub_key,".")
    if length(path) > 1 {
  	sub_key = replace!(path[-1],"_",".")
    }
    if get!(replacements, [sub_key]) != null {
  	sub_key = string!(get!(replacements, [sub_key]))
    } else {
  	sub_key = "system." + downcase(sub_key)
    }
    r.attributes = append(r.attributes,

        [{"key": sub_key, "value": {"stringValue": get!(.,path)}}]
    )
  }
  # Openshift object for grouping (dropped before sending)
  o = {
      "log_type": .log_type,
      "log_source": .log_source,
      "hostname": .hostname,
      "cluster_id": get!(.,["openshift","cluster_id"])
  }
  . = {
    "openshift": o,
    "resource": resource,
    "logRecords": r
  }
'''

# Normalize audit log record to OTLP semantic conventions
[transforms.output_otel_collector_otlp_audit_linux]
type = "remap"
inputs = ["output_otel_collector_reroute.linux"]
source = '''
  # Create base resource attributes
  resource.attributes = []
  resource.attributes = append( resource.attributes, 
      [{"key": "node.name", "value": {"stringValue": .hostname}},
      {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
  )
  # Create logRecord object
  r = {}
  r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
  r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
  # Convert syslog severity keyword to number, default to 9 (unknown)
  r.severityNumber = to_syslog_severity(.level) ?? 9
  r.body = {"stringValue": string!(.message)}
  r.attributes = []
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "openshift.log.type", "value": {"stringValue": .log_type}},
  	{"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
  )
  # Openshift object for grouping (dropped before sending)
  o = {
      "log_type": .log_type,
      "log_source": .log_source,
      "hostname": .hostname,
      "cluster_id": get!(.,["openshift","cluster_id"])
  }
  . = {
    "openshift": o,
    "resource": resource,
    "logRecords": r
  }
'''

# Normalize audit log kube record to OTLP semantic conventions
[transforms.output_otel_collector_otlp_audit_kube]
type = "remap"
inputs = ["output_otel_collector_reroute.kube"]
source = '''
  # Create base resource attributes
  resource.attributes = []
  resource.attributes = append( resource.attributes, 
      [{"key": "node.name", "value": {"stringValue": .hostname}},
      {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
  )
  # Create logRecord object
  r = {}
  r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
  r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
  # Convert syslog severity keyword to number, default to 9 (unknown)
  r.severityNumber = to_syslog_severity(.level) ?? 9
  r.body = {"stringValue": string!(.message)}
  r.attributes = []
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "openshift.log.type", "value": {"stringValue": .log_type}},
  	{"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
  )
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "url.full", "value": {"stringValue": .requestURI}},
  	{"key": "http.response.status.code", "value": {"stringValue": to_string!(get!(.,["responseStatus","code"]))}},
  	{"key": "http.request.method", "value": {"stringValue": .verb}}]
  )
  # Openshift object for grouping (dropped before sending)
  o = {
      "log_type": .log_type,
      "log_source": .log_source,
      "hostname": .hostname,
      "cluster_id": get!(.,["openshift","cluster_id"])
  }
  . = {
    "openshift": o,
    "resource": resource,
    "logRecords": r
  }
'''

# Normalize audit openshiftAPI record to OTLP semantic conventions
[transforms.output_otel_collector_otlp_audit_openshift]
type = "remap"
inputs = ["output_otel_collector_reroute.openshift"]
source = '''
  # Create base resource attributes
  resource.attributes = []
  resource.attributes = append( resource.attributes, 
      [{"key": "node.name", "value": {"stringValue": .hostname}},
      {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
  )
  # Create logRecord object
  r = {}
  r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
  r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
  # Convert syslog severity keyword to number, default to 9 (unknown)
  r.severityNumber = to_syslog_severity(.level) ?? 9
  r.body = {"stringValue": string!(.message)}
  r.attributes = []
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "openshift.log.type", "value": {"stringValue": .log_type}},
  	{"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
  )
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "url.full", "value": {"stringValue": .requestURI}},
  	{"key": "http.response.status.code", "value": {"stringValue": to_string!(get!(.,["responseStatus","code"]))}},
  	{"key": "http.request.method", "value": {"stringValue": .verb}}]
  )
  # Openshift object for grouping (dropped before sending)
  o = {
      "log_type": .log_type,
      "log_source": .log_source,
      "hostname": .hostname,
      "cluster_id": get!(.,["openshift","cluster_id"])
  }
  . = {
    "openshift": o,
    "resource": resource,
    "logRecords": r
  }
'''

# Normalize audit log ovn records to OTLP semantic conventions
[transforms.output_otel_collector_otlp_audit_ovn]
type = "remap"
inputs = ["output_otel_collector_reroute.ovn"]
source = '''
  # Create base resource attributes
  resource.attributes = []
  resource.attributes = append( resource.attributes, 
      [{"key": "node.name", "value": {"stringValue": .hostname}},
      {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
  )
  # Create logRecord object
  r = {}
  r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
  r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
  # Convert syslog severity keyword to number, default to 9 (unknown)
  r.severityNumber = to_syslog_severity(.level) ?? 9
  r.body = {"stringValue": string!(.message)}
  r.attributes = []
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "openshift.log.type", "value": {"stringValue": .log_type}},
  	{"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
  )
  # Append logRecord attributes
  r.attributes = append(
  	r.attributes,
  	[{"key": "url.full", "value": {"stringValue": .requestURI}},
  	{"key": "http.response.status.code", "value": {"stringValue": to_string!(get!(.,["responseStatus","code"]))}},
  	{"key": "http.request.method", "value": {"stringValue": .verb}}]
  )
  # Openshift object for grouping (dropped before sending)
  o = {
      "log_type": .log_type,
      "log_source": .log_source,
      "hostname": .hostname,
      "cluster_id": get!(.,["openshift","cluster_id"])
  }
  . = {
    "openshift": o,
    "resource": resource,
    "logRecords": r
  }
'''

# Merge audit and node logs and group by hostname and log_type
[transforms.output_otel_collector_group_by_source]
type = "reduce"
inputs = ["output_otel_collector_otlp_audit_kube","output_otel_collector_otlp_audit_linux","output_otel_collector_otlp_audit_openshift","output_otel_collector_otlp_audit_ovn","output_otel_collector_otlp_node"]
expire_after_ms = 10000
max_events = 3
group_by = [".openshift.cluster_id",".openshift.hostname",".openshift.log_type"]
merge_strategies.resource = "retain"
merge_strategies.logRecords = "array"

# Create new resource object for OTLP JSON payload
[transforms.output_otel_collector_final_otlp]
type = "remap"
inputs = ["output_otel_collector_group_by_container","output_otel_collector_group_by_source","output_otel_collector_reroute._unmatched"]
source = '''
  . = {
        "resource": {
           "attributes": .resource.attributes,
        },
        "scopeLogs": [
          {"logRecords": .logRecords}
        ]
      }
'''

[sinks.output_otel_collector]
type = "http"
inputs = ["output_otel_collector_final_otlp"]
uri = "http://localhost:4318/v1/logs"
method = "post"
payload_prefix = "{\"resourceLogs\":"
payload_suffix = "}"
encoding.codec = "json"

`,
		),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
