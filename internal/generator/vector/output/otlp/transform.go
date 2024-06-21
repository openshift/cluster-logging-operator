package otlp

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

// VRL for OTLP transforms by route
const (
	CreateResourceAttributes = `
# Create base resource attributes
resource.attributes = []
resource.attributes = append( resource.attributes, 
    [{"key": "node.name", "value": {"stringValue": .hostname}},
    {"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}}]
)
`
	AppendContainerAttributes = `
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
`
	CreateLogRecord = `
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

`
	AppendAPILogRecordAttributes = `
# Append logRecord attributes
r.attributes = append(
	r.attributes,
	[{"key": "url.full", "value": {"stringValue": .requestURI}},
	{"key": "http.response.status.code", "value": {"stringValue": to_string!(get!(.,["responseStatus","code"]))}},
	{"key": "http.request.method", "value": {"stringValue": .verb}}]
)
`
	AppendNodeLogRecordAttributes = `
# Append log attributes for node logs
logAttribute = [
  "systemd.t.BOOT_ID",
  "systemd.t.CMDLINE",
  "systemd.t.EXE",
  "systemd.t.GID",
  "systemd.t.MACHINE_ID",
  "systemd.t.PID",
  "systemd.u.SYSLOG_FACILITY",
  "systemd.u.SYSLOG_IDENTIFIER",
]
replacements = {
  "SYSLOG.FACILITY": "syslog.facility",
  "SYSLOG.IDENTIFIER": "syslog.identifier",
  "PID": "syslog.procid"
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
`
	FinalObjectWithOpenshiftGrouping = `
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
`
	FinalObjectWithKubeContainerGrouping = `
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
`
)

func containerLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		CreateResourceAttributes,
		AppendContainerAttributes,
		CreateLogRecord,
		FinalObjectWithKubeContainerGrouping,
	}), "\n")
}

func nodeLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		CreateResourceAttributes,
		CreateLogRecord,
		AppendNodeLogRecordAttributes,
		FinalObjectWithOpenshiftGrouping,
	}), "\n")
}

func auditHostLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		CreateResourceAttributes,
		CreateLogRecord,
		FinalObjectWithOpenshiftGrouping,
	}), "\n")
}

func auditAPILogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		CreateResourceAttributes,
		CreateLogRecord,
		AppendAPILogRecordAttributes,
		FinalObjectWithOpenshiftGrouping,
	}), "\n")
}

func TransformContainer(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Normalize container log records to OTLP semantic conventions",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         containerLogsVRL(),
	}
}

func TransformJournal(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Normalize node log events to OTLP semantic conventions",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         nodeLogsVRL(),
	}
}

func TransformAuditHost(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Normalize audit log record to OTLP semantic conventions",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         auditHostLogsVRL(),
	}
}

func TransformAuditKube(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Normalize audit log kube record to OTLP semantic conventions",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         auditAPILogsVRL(),
	}
}
func TransformAuditOpenshift(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Normalize audit openshiftAPI record to OTLP semantic conventions",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         auditAPILogsVRL(),
	}
}
func TransformAuditOvn(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Normalize audit log ovn records to OTLP semantic conventions",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         auditAPILogsVRL(),
	}
}

// FormatResourceLog Drops everything except resource.attributes and scopeLogs.logRecords
func FormatResourceLog(id string, inputs []string) Element {
	return elements.Remap{
		Desc:        "Create new resource object for OTLP JSON payload",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL: strings.TrimSpace(`
. = {
      "resource": {
         "attributes": .resource.attributes,
      },
      "scopeLogs": [
        {"logRecords": .logRecords}
      ]
    }
`),
	}
}
