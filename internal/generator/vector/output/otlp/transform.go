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
resource.attributes = append( resource.attributes, 
    [{"key": "k8s.cluster.uid", "value": {"stringValue": get!(.,["openshift","cluster_id"])}},
    {"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
)
`
	HostResourceAttributes = `
# Append auditd host attributes
resource.attributes = append( resource.attributes,
    [{"key": "k8s.node.name", "value": {"stringValue": .hostname}}]
)
`
	ContainerResourceAttributes = `
# Append container resource attributes
resource.attributes = append( resource.attributes,
    [{"key": "k8s.pod.name", "value": {"stringValue": get!(.,["kubernetes","pod_name"])}},
    {"key": "k8s.container.name", "value": {"stringValue": get!(.,["kubernetes","container_name"])}},
    {"key": "k8s.namespace.name", "value": {"stringValue": get!(.,["kubernetes","namespace_name"])}}]
)
`
	LogRecord = `
# Create logRecord object
r = {}
r.timeUnixNano = to_string(to_unix_timestamp(parse_timestamp!(.@timestamp, format:"%+"), unit:"nanoseconds"))
r.observedTimeUnixNano = to_string(to_unix_timestamp(now(), unit:"nanoseconds"))
# Convert syslog severity keyword to number, default to 9 (unknown)
r.severityNumber = to_syslog_severity(.level) ?? 9
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
	LogAttributes = `
# Create logRecord attributes
r.attributes = []
r.attributes = append(r.attributes,
    [{"key": "openshift.log.type", "value": {"stringValue": .log_type}}]
)
if exists(.openshift.labels) {for_each(object!(.openshift.labels)) -> |key,value| {
    r.attributes = append(r.attributes,
        [{"key": "openshift.label." + key, "value": {"stringValue": value}}]
    )
}}
`
	ContainerLogAttributes = `
# Append kube pod labels
r.attributes = append(r.attributes,
    [{"key": "k8s.pod.uid", "value": {"stringValue": get!(.,["kubernetes","pod_id"])}},
    {"key": "k8s.container.id", "value": {"stringValue": get!(.,["kubernetes","container_id"])}},
    {"key": "k8s.node.name", "value": {"stringValue": .hostname}},
    {"key": "log.iostream", "value": {"stringValue": get!(.,["kubernetes","container_iostream"])}}]
)
if exists(.kubernetes.labels) {for_each(object!(.kubernetes.labels)) -> |key,value| {
    r.attributes = append(r.attributes,
        [{"key": "k8s.pod.label." + key, "value": {"stringValue": value}}]
    )
}}
`
	APILogAttributes = `
# Append API logRecord attributes
parts = split(to_string!(.requestURI), "?")
r.attributes = append(r.attributes,
	[{"key": "http.response.status.code", "value": {"stringValue": to_string!(get!(.,["responseStatus","code"]))}},
	{"key": "http.request.method_original", "value": {"stringValue": .verb}},
    {"key": "user.name", "value": {"stringValue": get!(.,["user","username"])}},
    {"key": "user_agent.original", "value": {"stringValue": .userAgent }},
    {"key": "url.domain", "value": {"stringValue": .hostname }},
	{"key": "url.path", "value": {"stringValue": parts[0] }},
	{"key": "url.query", "value": {"stringValue": parts[1] }}]
)
`
	NodeLogAttributes = `
# Append log attributes for node logs
r.attributes = append(r.attributes,
	[{"key": "syslog.facility", "value": {"stringValue": to_string!(get!(.,["systemd","u","SYSLOG_FACILITY"]))}},
	{"key": "service.name", "value": {"stringValue": to_string!(get!(.,["systemd","u","SYSLOG_IDENTIFIER"]))}},
	{"key": "process.command", "value": {"stringValue": to_string!(get!(.,["systemd","t","COMM"]))}},
	{"key": "process.command_line", "value": {"stringValue": to_string!(get!(.,["systemd","t","CMDLINE"]))}},
	{"key": "process.executable.path", "value": {"stringValue": to_string!(get!(.,["systemd","t","EXE"]))}},
	{"key": "process.gid", "value": {"stringValue": to_string!(get!(.,["systemd","t","GID"]))}},
	{"key": "host.id", "value": {"stringValue": to_string!(get!(.,["systemd","t","MACHINE_ID"]))}},
    {"key": "host.name", "value": {"stringValue": .hostname}},
	{"key": "process.pid", "value": {"stringValue": to_string!(get!(.,["systemd","t","PID"]))}},
	{"key": "process.user.id", "value": {"stringValue": to_string!(get!(.,["systemd","t","UID"]))}}]
)
`
	FinalGrouping = `
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
	FinalGroupingContainers = `
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
	// The (M)inimum (V)iable (P)roduct Labels (MVP)
	BackwardCompatBaseResourceAttributes = `
# Append backward compatibility attributes
resource.attributes = append( resource.attributes,
	[{"key": "log_type", "value": {"stringValue": .log_type}}]
)
`
	BackwardCompatContainerResourceAttributes = `
# Append backward compatibility attributes for container logs
resource.attributes = append( resource.attributes,
	[{"key": "kubernetes_pod_name", "value": {"stringValue": get!(.,["kubernetes","pod_name"])}},
	{"key": "kubernetes_container_name", "value": {"stringValue": get!(.,["kubernetes","container_name"])}},
	{"key": "kubernetes_namespace_name", "value": {"stringValue": get!(.,["kubernetes","namespace_name"])}}]
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
		LogRecord,
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
		HostResourceAttributes,
		LogRecord,
		BodyFromInternal,
		LogAttributes,
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

func auditOvnLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		BackwardCompatBaseResourceAttributes,
		LogRecord,
		BodyFromMessage,
		LogAttributes,
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
		VRL:         auditOvnLogsVRL(),
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
