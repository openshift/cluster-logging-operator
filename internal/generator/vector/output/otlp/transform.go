package otlp

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

// VRL for OTLP transforms by route
const (
	BaseResourceAttributes = `
# Create base resource attributes
resource.attributes = []
resource.attributes = append( resource.attributes, 
    [{"key": "cluster.id", "value": {"stringValue": get!(.,["openshift","cluster_id"])}},
    {"key": "openshift.log.source", "value": {"stringValue": .log_source}}]
)
if exists(.openshift.sequence) {
	resource.attributes = append(resource.attributes,[{"key": "openshift.sequence", "value": {"intValue": .openshift.sequence}}])
}
`
	HostResourceAttributes = `
# Append auditd host attributes
resource.attributes = append( resource.attributes,
    [{"key": "node.name", "value": {"stringValue": .hostname}}]
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
    {"key": "k8s.container.id", "value": {"stringValue": get!(.,["kubernetes","container_id"])}},]
)
if exists(.kubernetes.labels) {for_each(object!(.kubernetes.labels)) -> |key,value| {
    r.attributes = append(r.attributes,
        [{"key": "k8s.pod.label." + key, "value": {"stringValue": value}}]
    )
}}
`
	APILogAttributes = `
# Append API logRecord attributes
r.attributes = append(r.attributes,
	[{"key": "url.full", "value": {"stringValue": .requestURI}},
	{"key": "http.response.status.code", "value": {"stringValue": to_string!(get!(.,["responseStatus","code"]))}},
	{"key": "http.request.method", "value": {"stringValue": .verb}}]
)
`
	NodeLogAttributes = `
# Append log attributes for node logs
r.attributes = append(r.attributes,
	[{"key": "syslog.facility", "value": {"stringValue": to_string!(get!(.,["systemd","u","SYSLOG_FACILITY"]))}},
	{"key": "syslog.identifier", "value": {"stringValue": to_string!(get!(.,["systemd","u","SYSLOG_IDENTIFIER"]))}},
	{"key": "syslog.procid", "value": {"stringValue": to_string!(get!(.,["systemd","t","PID"]))}},
	{"key": "system.unit", "value": {"stringValue": to_string!(get!(.,["systemd","t","SYSTEMD_UNIT"]))}},
	{"key": "system.uid", "value": {"stringValue": to_string!(get!(.,["systemd","t","UID"]))}},
	{"key": "system.slice", "value": {"stringValue": to_string!(get!(.,["systemd","t","SYSTEMD_SLICE"]))}},
	{"key": "system.cgroup", "value": {"stringValue": to_string!(get!(.,["systemd","t","SYSTEMD_CGROUP"]))}},
	{"key": "system.cmdline", "value": {"stringValue": to_string!(get!(.,["systemd","t","CMDLINE"]))}},
	{"key": "system.invocation.id", "value": {"stringValue": to_string!(get!(.,["systemd","t","SYSTEMD_INVOCATION_ID"]))}}]
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
)

func containerLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		BaseResourceAttributes,
		ContainerResourceAttributes,
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
