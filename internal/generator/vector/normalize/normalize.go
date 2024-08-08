package normalize

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"

	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	FixLogLevel = `
if !exists(.level) {
  .level = "default"

  # Match on well known structured patterns
  # Order: emergency, alert, critical, error, warn, notice, info, debug

  if match!(.message, r'^EM[0-9]+|level=emergency|Value:emergency|"level":"emergency"') {
    .level = "emergency"
  } else if match!(.message, r'^A[0-9]+|level=alert|Value:alert|"level":"alert"') {
    .level = "alert"
  } else if match!(.message, r'^C[0-9]+|level=critical|Value:critical|"level":"critical"') {
    .level = "critical"
  } else if match!(.message, r'^E[0-9]+|level=error|Value:error|"level":"error"') {
    .level = "error"
  } else if match!(.message, r'^W[0-9]+|level=warn|Value:warn|"level":"warn"') {
    .level = "warn"
  } else if match!(.message, r'^N[0-9]+|level=notice|Value:notice|"level":"notice"') {
    .level = "notice"
  } else if match!(.message, r'^I[0-9]+|level=info|Value:info|"level":"info"') {
    .level = "info"
  } else if match!(.message, r'^D[0-9]+|level=debug|Value:debug|"level":"debug"') {
    .level = "debug"
  }

  # Match on unstructured keywords in same order

  if .level == "default" {
    if match!(.message, r'Emergency|EMERGENCY|<emergency>') {
      .level = "emergency"
    } else if match!(.message, r'Alert|ALERT|<alert>') {
      .level = "alert"
    } else if match!(.message, r'Critical|CRITICAL|<critical>') {
      .level = "critical"
    } else if match!(.message, r'Error|ERROR|<error>') {
      .level = "error"
    } else if match!(.message, r'Warning|WARN|<warn>') {
      .level = "warn"
    } else if match!(.message, r'Notice|NOTICE|<notice>') {
      .level = "notice"
    } else if match!(.message, r'(?i)\b(?:info)\b|<info>') {
      .level = "info"
    } else if match!(.message, r'Debug|DEBUG|<debug>') {
      .level = "debug"
    }
  }
}
`
	RemoveSourceType     = `del(.source_type)`
	HandleEventRouterLog = `
pod_name = string!(.kubernetes.pod_name)
if starts_with(pod_name, "eventrouter-") {
  parsed, err = parse_json(.message)
  if err != null {
    log("Unable to process EventRouter log: " + err, level: "info")
  } else {
    ., err = merge(.,parsed)
    if err == null && exists(.event) && is_object(.event) {
        if exists(.verb) {
          .event.verb = .verb
          del(.verb)
        }
        .kubernetes.event = del(.event)
        .message = del(.kubernetes.event.message)
        set!(., ["@timestamp"], .kubernetes.event.metadata.creationTimestamp)
        del(.kubernetes.event.metadata.creationTimestamp)
		. = compact(., nullish: true)
    } else {
      log("Unable to merge EventRouter log message into record: " + err, level: "info")
    }
  }
}
`
	RemoveStream       = `del(.stream)`
	RemovePodIPs       = `del(.kubernetes.pod_ips)`
	RemoveNodeLabels   = `del(.kubernetes.node_labels)`
	RemoveTimestampEnd = `del(.timestamp_end)`

	ParseHostAuditLogs = `
match1 = parse_regex(.message, r'type=(?P<type>[^ ]+)') ?? {}
envelop = {}
envelop |= {"type": match1.type}

match2, err = parse_regex(.message, r'msg=audit\((?P<ts_record>[^ ]+)\):')
if err == null {
  sp, err = split(match2.ts_record,":")
  if err == null && length(sp) == 2 {
      ts = parse_timestamp(sp[0],"%s.%3f") ?? ""
      envelop |= {"record_id": sp[1]}
      . |= {"audit.linux" : envelop}
      . |= {"@timestamp" : format_timestamp(ts,"%+") ?? ""}
  }
} else {
  log("could not parse host audit msg. err=" + err, rate_limit_secs: 0)
}
`
	HostAuditLogTag = ".linux-audit.log"
	K8sAuditLogTag  = ".k8s-audit.log"
	OpenAuditLogTag = ".openshift-audit.log"
	OvnAuditLogTag  = ".ovn-audit.log"
	ParseAndFlatten = `. = merge(., parse_json!(string!(.message))) ?? .
del(.message)
`
	FixHostname = `.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`

	FixK8sAuditLevel       = `.k8s_audit_level = .level`
	FixOpenshiftAuditLevel = `.openshift_audit_level = .level`
	AddDefaultLogLevel     = `.level = "default"`
)

var (
	AddHostAuditTag = fmt.Sprintf(".tag = %q", HostAuditLogTag)
	AddK8sAuditTag  = fmt.Sprintf(".tag = %q", K8sAuditLogTag)
	AddOpenAuditTag = fmt.Sprintf(".tag = %q", OpenAuditLogTag)
	AddOvnAuditTag  = fmt.Sprintf(".tag = %q", OvnAuditLogTag)
)

func NormalizeContainerLogs(inputs, id string) []framework.Element {
	return []framework.Element{
		Remap{
			ComponentID: id,
			Inputs:      helpers.MakeInputs(inputs),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				ClusterID,
				FixLogLevel,
				HandleEventRouterLog,
				RemoveSourceType,
				RemoveStream,
				RemovePodIPs,
				RemoveNodeLabels,
				RemoveTimestampEnd,
				FixTimestampField,
			}), "\n"),
		},
	}
}

func NormalizeHostAuditLogs(inLabel, outLabel string) []framework.Element {
	return []framework.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				ClusterID,
				AddHostAuditTag,
				ParseHostAuditLogs,
				AddDefaultLogLevel,
			}), "\n\n"),
		},
	}
}

func NormalizeK8sAuditLogs(inputs, id string) []framework.Element {
	return []framework.Element{
		Remap{
			ComponentID: id,
			Inputs:      helpers.MakeInputs(inputs),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				ClusterID,
				AddK8sAuditTag,
				ParseAndFlatten,
				FixK8sAuditLevel,
			}), "\n"),
		},
	}
}

func NormalizeOpenshiftAuditLogs(inLabel, outLabel string) []framework.Element {
	return []framework.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				ClusterID,
				AddOpenAuditTag,
				ParseAndFlatten,
				FixOpenshiftAuditLevel,
			}), "\n"),
		},
	}
}

func NormalizeOVNAuditLogs(inLabel, outLabel string) []framework.Element {
	return []framework.Element{
		Remap{
			ComponentID: outLabel,
			Inputs:      helpers.MakeInputs(inLabel),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				ClusterID,
				AddOvnAuditTag,
				FixLogLevel,
			}), "\n"),
		},
	}
}
