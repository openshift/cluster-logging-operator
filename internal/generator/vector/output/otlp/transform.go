package otlp

import (
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

// VRL for OTLP transforms by route
const (
	BaseResourceAttributes = `
# Create base resource attributes
resource.attributes = []
resource.attributes = append(resource.attributes,
  [
    {"key": "openshift.cluster.uid", "value": {"stringValue": .openshift.cluster_id}},
    {"key": "openshift.log.source", "value": {"stringValue": .log_source}},
    {"key": "openshift.log.type", "value": {"stringValue": .log_type}},
    {"key": "k8s.node.name", "value": {"stringValue": .hostname}}
  ]
)
if exists(.openshift.labels) {for_each(object!(.openshift.labels)) -> |key,value| {
    resource.attributes = append(resource.attributes,
        [{"key": "openshift.label." + key, "value": {"stringValue": value}}]
    )
}}
`
	ContainerResourceAttributes = `
resource.attributes = append( resource.attributes,
  [
    {"key": "k8s.pod.name", "value": {"stringValue": .kubernetes.pod_name}},
	{"key": "k8s.pod.uid", "value": {"stringValue": .kubernetes.pod_id}},
    {"key": "k8s.container.name", "value": {"stringValue": .kubernetes.container_name}},
    {"key": "k8s.namespace.name", "value": {"stringValue": .kubernetes.namespace_name}}
  ]
)
if exists(.kubernetes.labels) {for_each(object!(.kubernetes.labels)) -> |key,value| {
    resource.attributes = append(resource.attributes,
        [{"key": "k8s.pod.label." + key, "value": {"stringValue": value}}]
    )
}}
`
	LogRecord = `
# Create logRecord object
r = {"attributes": []}
r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
`
	LogRecordSeverity = `
r.severityText = .level
`
	BodyFromMessage = `
# Create body from original message or structured
value = .message
if (value == null) { value = encode_json(.structured) }
r.body = {"stringValue": string!(value)}
`
	BodyFromInternal = `
# Create body from internal message
r.body = {"stringValue": to_string!(get!(.,["_internal","message"]))}
`
	APILogAttributes = `
r.attributes = append(r.attributes,
  [
    {"key": "k8s.audit.event.level", "value": {"stringValue": .level}},
    {"key": "k8s.audit.event.stage", "value": {"stringValue": .stage}},
    {"key": "k8s.audit.event.request.uri", "value": {"stringValue": .requestURI}},
    {"key": "k8s.audit.event.request.verb", "value": {"stringValue": .verb}},
    {"key": "k8s.audit.event.user_agent", "value": {"stringValue": .userAgent}},
    {"key": "k8s.user.username", "value": {"stringValue": .user.username}}
  ]
)
if exists(.responseStatus.code) {
  r.attributes = push(r.attributes,{"key": "k8s.audit.event.response.code", "value": {"intValue": to_string!(.responseStatus.code)}})  
}
values = []
for_each(array!(.user.groups)) -> |_index,group| {
    .group = group
    values = push(values,{"stringValue": group})
}
r.attributes = push(r.attributes,{"key": "k8s.user.groups", "value": {"arrayValue": {"values": values}}})
if exists(.objectRef) {
  r.attributes = append(r.attributes,[
      {"key": "k8s.audit.event.object_ref.resource", "value": {"stringValue": .objectRef.resource}},
      {"key": "k8s.audit.event.object_ref.name", "value": {"stringValue": .objectRef.name}},
      {"key": "k8s.audit.event.object_ref.namespace", "value": {"stringValue": .objectRef.namespace}},
      {"key": "k8s.audit.event.object_ref.api_version", "value": {"stringValue": .objectRef.apiVersion}},
      {"key": "k8s.audit.event.object_ref.api_group", "value": {"stringValue": .objectRef.apiGroup}}
    ]
  )
}
if exists(.annotations) {for_each(object!(.annotations)) -> |key,value| {
    r.attributes = append(r.attributes,
        [{"key": "k8s.audit.event.annotation." + key, "value": {"stringValue": value}}]
    )
}}
`
	OVNLogAttributes = `
# Fill up OVN logRecord object
if exists(.level) { r.severityText = .level } 
ovnTokens = split(to_string!(get!(.,["_internal","message"])),"|")
if 0 < length(ovnTokens) { r.attributes = push(r.attributes, {"key": "k8s.ovn.sequence", "value": {"stringValue": ovnTokens[1] }})}
if 1 < length(ovnTokens) { r.attributes = push(r.attributes, {"key": "k8s.ovn.component", "value": {"stringValue": ovnTokens[2] }})}
`
	HostLogAttributes = `
# Fill up auditd logRecord object
if exists(.level) { r.severityText = .level } 
kv = parse_key_value!(to_string!(get!(.,["_internal","message"])))
if exists(kv.type) {
  r.attributes = push(r.attributes, {"key": "auditd.type", "value": {"stringValue": kv.type }})
}
if exists(kv.msg) {
  trimmed = slice!(kv.msg, find!(kv.msg, "(") + 1, -2)
  parts = split!(trimmed, ":")
  r.attributes = push(r.attributes, {"key": "auditd.sequence", "value": {"stringValue": parts[1] }})
}
`

	LogAttributes          = ``
	ContainerLogAttributes = `
r.attributes = append(r.attributes,
  [
	{"key": "log.iostream", "value": {"stringValue": .kubernetes.container_iostream}},
	{"key": "level", "value": {"stringValue": .level}}
  ]
)
`
	NodeResourceAttributes = `
resource.attributes = append(resource.attributes,
  [
	{"key": "process.command_line", "value": {"stringValue": .systemd.t.CMDLINE}},
	{"key": "process.executable.name", "value": {"stringValue": .systemd.t.COMM}},
	{"key": "process.executable.path", "value": {"stringValue": .systemd.t.EXE}},
	{"key": "process.pid", "value": {"stringValue": .systemd.t.PID}},
	{"key": "service.name", "value": {"stringValue": .systemd.t.SYSTEMD_UNIT}}
  ]
)
`
	NodeLogAttributes = `
r.attributes = append(r.attributes, [{"key": "level", "value": {"stringValue": .level}}])
if exists(.systemd.t) {for_each(object!(.systemd.t)) -> |key,value| {
    r.attributes = append(r.attributes,
        [{"key": "systemd.t." + downcase(key), "value": {"stringValue": value}}]
    )
}}
if exists(.systemd.u) {for_each(object!(.systemd.u)) -> |key,value| {
    r.attributes = append(r.attributes,
        [{"key": "systemd.u." + downcase(key), "value": {"stringValue": value}}]
    )
}}
`
	FinalGrouping = `
# Openshift object for grouping (dropped before sending)
o = {
    "log_type": .log_type,
    "log_source": .log_source,
    "hostname": .hostname,
    "cluster_id": .openshift.cluster_id
}
. = {
  "openshift": o,
  "resource": resource,
  "logRecords": r
}
`
	FinalGroupingContainers = `
# Openshift and kubernetes objects for grouping containers (dropped before sending)
o = {
    "log_type": .log_type,
    "log_source": .log_source,
    "cluster_id": .openshift.cluster_id
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
	// The (M)inimum (V)iable (P)roduct Labels (MVP)
	BackwardCompatBaseResourceAttributes = `
# Append backward compatibility attributes
resource.attributes = append( resource.attributes,
	[
      {"key": "log_type", "value": {"stringValue": .log_type}},
      {"key": "log_source", "value": {"stringValue": .log_source}},
      {"key": "openshift.cluster_id", "value": {"stringValue": .openshift.cluster_id}},
      {"key": "kubernetes.host", "value": {"stringValue": .hostname}}
    ]
)
`
	BackwardCompatContainerResourceAttributes = `
# Append backward compatibility attributes for container logs
resource.attributes = append( resource.attributes,
	[{"key": "kubernetes.pod_name", "value": {"stringValue": .kubernetes.pod_name}},
	{"key": "kubernetes.container_name", "value": {"stringValue": .kubernetes.container_name}},
	{"key": "kubernetes.namespace_name", "value": {"stringValue": .kubernetes.namespace_name}}]
)
`
)

func containerLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		ContainerResourceAttributes,
		BackwardCompatBaseResourceAttributes,
		BackwardCompatContainerResourceAttributes,
		LogRecord,
		LogRecordSeverity,
		BodyFromMessage,
		LogAttributes,
		ContainerLogAttributes,
		FinalGroupingContainers,
	}), "\n")
}

func nodeLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		BackwardCompatBaseResourceAttributes,
		NodeResourceAttributes,
		LogRecord,
		LogRecordSeverity,
		BodyFromMessage,
		LogAttributes,
		NodeLogAttributes,
		FinalGrouping,
	}), "\n")
}

func auditHostLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		BackwardCompatBaseResourceAttributes,
		LogRecord,
		BodyFromInternal,
		LogAttributes,
		HostLogAttributes,
		FinalGrouping,
	}), "\n")
}

func auditAPILogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		BackwardCompatBaseResourceAttributes,
		LogRecord,
		BodyFromInternal,
		LogAttributes,
		APILogAttributes,
		FinalGrouping,
	}), "\n")
}

func auditOVNLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		BackwardCompatBaseResourceAttributes,
		LogRecord,
		BodyFromInternal,
		LogAttributes,
		OVNLogAttributes,
		FinalGrouping,
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
		VRL:         auditOVNLogsVRL(),
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
