package syslog

import (
	"fmt"
	"net/url"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	TCP = `tcp`
	TLS = `tls`

	vrlKeySyslogFacility = "._syslog.facility"
	vrlKeySyslogSeverity = "._syslog.severity"
	vrlKeySyslogProcID   = "._syslog.proc_id"
	vrlKeySyslogAppName  = "._syslog.app_name"
	vrlKeySyslogMsgID    = "._syslog.msg_id"

	defProcIdRFC3164 = `to_string!(._syslog.proc_id || "")
if exists(._syslog.proc_id) && is_empty(strip_whitespace(string!(._syslog.proc_id))) { del(._syslog.proc_id) }
`
	defAppNameRFC3164 = `to_string!(._syslog.app_name || "")`

	// Default values for Syslog fields for infrastructure logType if source 'node'
	nodeAppNameRFC5424  = `._syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "-")`
	nodeAppNameRFC3164  = `._syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "")`
	nodeProcIdRFC3164   = `._syslog.proc_id = to_string!(.systemd.t.PID || "")`
	nodeProcIdRFC5424   = `._syslog.proc_id = to_string!(.systemd.t.PID || "-")`

	// Default values for Syslog fields for application logType and infrastructure logType if source 'container'
	containerAppNameRFC5424 = `._syslog.app_name, err = join([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "_")
if err != null {
  log("K8s metadata (namespace, pod, or container) missing; syslog.app_name set to '-'", level: "error")
  ._syslog.app_name = "-"
}
`
	containerAppNameRFC3164 = `._syslog.app_name, err = join([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "")
if err != null {
  log("K8s metadata (namespace, pod, or container) missing: unable to calculate syslog.app_name (TAG)", level: "error")
} else {
  #Remove non-alphanumeric characters
  ._syslog.app_name = replace(._syslog.app_name, r'[^a-zA-Z0-9]', "")
  #Truncate the sanitized tag to 32 characters
  ._syslog.app_name = truncate(._syslog.app_name, 32)
}
`
	containerFacility = `._syslog.facility = "user"`
	containerSeverity = `._syslog.severity = .level`
	containerProcId   = `._syslog.proc_id = to_string!(.kubernetes.pod_id || "")`

	// Default values for Syslog fields for audit logType
	auditAppName  = `._syslog.app_name = .log_source`
	auditSeverity = `._syslog.severity = "informational"`
	auditFacility = `._syslog.facility = "security"`
	auditProcId   = `._syslog.proc_id = to_string!(.auditID || "-")`

	msgId = `._syslog.msg_id = .log_source`

	// conditions
	isInfrastructureNodeLogCond = `.log_type == "infrastructure" && .log_source == "node"`
	isContainerLogCond          = `.log_source == "container"`
	isAuditLogCond              = `.log_type == "audit"`

	facilitySeverityConversionVRL = `
# try to convert syslog code to the facility, severity names (e.g. 4 -> "warning", "4" -> "warning"  )
if exists(._syslog.facility) {
  _, err = to_syslog_facility_code(._syslog.facility)
  if err != null {
    # Field is not a valid name — try treating it as a code (int or string int)
    code, err2 = to_int(._syslog.facility)
    if err2 == null {
      facility, err3 = to_syslog_facility(code)
      if err3 == null {
        ._syslog.facility = facility
      } else {
        log("Invalid syslog facility code: " + to_string!(._syslog.facility) , level: "warn")
      }
    } else {
      log("Invalid syslog facility value: " + to_string!(._syslog.facility), level: "warn")
    }
  }
  # else: already a valid name, leave it as-is
}

if exists(._syslog.severity) {
  _, err = to_syslog_severity(._syslog.severity)
  if err != null {
    # Field is not a valid name — try treating it as a code (int or string int)
    code, err2 = to_int(._syslog.severity)
    if err2 == null {
      severity, err3 = to_syslog_level(code)
      if err3 == null {
        ._syslog.severity = severity
      } else {
        log("Invalid syslog severity code: " + to_string!(._syslog.severity), level: "warn")
      }
    } else {
      log("Invalid syslog severity value: " + to_string!(._syslog.severity), level: "warn")
    }
  }
  # else: already a valid name, leave it as-is
}`

	payloadKeyDefaultVRL = `
# Payload key NOT configured, full payload set to .message field (skipping internal objects)
excluded_fields = ["_internal", "_syslog"]
temp = .
for_each(excluded_fields) -> |_index, field| {
  temp = remove(temp, [field]) ?? temp
}
.message = temp`

	payloadKeyConfiguredVRL = `
# Payload key configured, going to use the payload key for .message field
payload_key = %s
if payload_key != null && payload_key != "" {
  payload_key = string!(payload_key)
  path = split!(payload_key, ".")
  value, err = get(., path)
  if err == null && value != null {
    .message = value
  } else {
    log("payload_key not found in event, skipping", level: "warn")
  }
} else {
  excluded_fields = ["_internal", "_syslog"]
  temp = .
  for_each(excluded_fields) -> |_index, field| {
    temp = remove(temp, [field]) ?? temp
  }
  .message = temp
}`

	kubernetesMinimalEnrichmentVRL = `
# KubernetesMinimal
# Adds namespace_name, pod_name, and container_name to the beginning of the message body (e.g. namespace_name=myproject, container_name=server, pod_name=pod-123, message={"foo":"bar"}).
# This may result in the message body being an invalid JSON structure.

namespace = to_string(.kubernetes.namespace_name) ?? ""
container = to_string(.kubernetes.container_name) ?? ""
pod = to_string(.kubernetes.pod_name) ?? ""
msg_value = encode_json(.message)

.message = "namespace_name=\"" + namespace + "\"" +
           ", container_name=\"" + container + "\"" +
           ", pod_name=\"" + pod + "\"" +
           ", message=" + msg_value`
)

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	parseEncodingID := vectorhelpers.MakeID(id, "parse_encoding")
	tfs = api.Transforms{}
	tfs[parseEncodingID] = parseEncoding(inputs, o.Syslog)

	u, _ := url.Parse(o.Syslog.URL)
	mode := socketMode(u.Scheme)
	sink = sinks.NewSocket(mode, func(s *sinks.Socket) {
		s.Address = u.Host
		if mode == sinks.SocketModeTCP {
			s.Keepalive = &sinks.Keepalive{
				TimeSecs: 60,
			}
		}
		s.Encoding = buildSocketEncoding(o.OutputSpec)
		s.TLS = tls.NewTlsEnabled(o, secrets, op)
		s.Buffer = common.NewApiBuffer(o)
	}, parseEncodingID)
	return id, sink, tfs
}

func socketMode(urlScheme string) sinks.SocketMode {
	switch strings.ToLower(urlScheme) {
	case TCP, TLS:
		return sinks.SocketModeTCP
	case "udp":
		return sinks.SocketModeUDP
	default:
		return sinks.SocketMode(strings.ToLower(urlScheme))
	}
}

func buildSocketEncoding(o obs.OutputSpec) *sinks.SocketEncoding {
	syslogConfig := &sinks.SyslogEncodingConfig{
		RFC:      strings.ToLower(string(o.Syslog.RFC)),
		Facility: "._syslog.facility",
		Severity: "._syslog.severity",
		AppName:  "._syslog.app_name",
		ProcID:   "._syslog.proc_id",
	}

	if o.Syslog.RFC == obs.SyslogRFC5424 {
		syslogConfig.MsgID = "._syslog.msg_id"
	}

	return &sinks.SocketEncoding{
		Codec:  "syslog",
		Syslog: syslogConfig,
	}
}

func parseEncoding(inputs []string, o *obs.Syslog) types.Transform {
	vrls := []string{"\n._syslog = {}", buildDefaults(o)}

	appendField := func(field, userValue, defaultValue string) {
		vrl := defaultValue
		if userValue != "" {
			vrl = commontemplate.TransformUserTemplateToVRL(userValue)
		}
		if vrl != "" {
			vrls = append(vrls, fmt.Sprintf("%s = %s", field, vrl))
		}
	}

	appendField(vrlKeySyslogFacility, o.Facility, "")
	appendField(vrlKeySyslogSeverity, o.Severity, "")

	if o.RFC == obs.SyslogRFC3164 {
		appendField(vrlKeySyslogProcID, o.ProcId, defProcIdRFC3164)
		appendField(vrlKeySyslogAppName, o.AppName, defAppNameRFC3164)
	} else {
		appendField(vrlKeySyslogProcID, o.ProcId, "")
		appendField(vrlKeySyslogAppName, o.AppName, "")
		appendField(vrlKeySyslogMsgID, o.MsgId, "")
	}

	vrls = append(vrls, facilitySeverityConversionVRL)

	if key := PayloadKey(o.PayloadKey); key != "" {
		vrls = append(vrls, fmt.Sprintf(payloadKeyConfiguredVRL, key))
	} else {
		vrls = append(vrls, payloadKeyDefaultVRL)
	}

	if o.Enrichment == obs.EnrichmentTypeKubernetesMinimal {
		vrls = append(vrls, kubernetesMinimalEnrichmentVRL)
	}

	return transforms.NewRemap(strings.Join(vrls, "\n"), inputs...)
}

// PayloadKey returns the field path from a user template like "{.plKey}" -> ".plKey".
// Returns empty string if no payload key is configured.
func PayloadKey(plKey string) string {
	if plKey == "" {
		return ""
	}
	return plKey[1 : len(plKey)-1]
}

// defaultRule defines the structure for syslog default configuration rules
type defaultRule struct {
	cond     string
	appName  string
	procId   string
	severity string
	facility string
	msgId    string
}

// buildDefaults generates syslog default rules based on the provided configuration
func buildDefaults(o *obs.Syslog) string {
	if o == nil {
		return ""
	}

	var builder strings.Builder
	var rules []defaultRule

	switch o.RFC {
	case obs.SyslogRFC3164:
		buildRFC3164Rules(o, &rules)
	case obs.SyslogRFC5424:
		buildRFC5424Rules(o, &rules)
	}

	for _, rule := range rules {
		if rule.cond != "" {
			fmt.Fprintf(&builder, "if %s {\n", rule.cond)
		}
		for _, s := range []string{rule.appName, rule.procId, rule.severity, rule.facility, rule.msgId} {
			if s != "" {
				builder.WriteString(s + "\n")
			}
		}
		if rule.cond != "" {
			builder.WriteString("}\n")
		}
	}

	return builder.String()
}

func defaultIfEmpty(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return ""
}

// buildRFC3164Rules constructs the default rules for RFC3164 format
func buildRFC3164Rules(o *obs.Syslog, rules *[]defaultRule) {
	if o.ProcId == "" || o.AppName == "" {
		*rules = append(*rules, defaultRule{
			cond:    isInfrastructureNodeLogCond,
			appName: defaultIfEmpty(o.AppName, nodeAppNameRFC3164),
			procId:  defaultIfEmpty(o.ProcId, nodeProcIdRFC3164),
		})
	}

	if o.AppName == "" || o.Severity == "" || o.Facility == "" {
		*rules = append(*rules,
			defaultRule{
				cond:     isContainerLogCond,
				appName:  defaultIfEmpty(o.AppName, containerAppNameRFC3164),
				severity: defaultIfEmpty(o.Severity, containerSeverity),
				facility: defaultIfEmpty(o.Facility, containerFacility),
			},
			defaultRule{
				cond:     isAuditLogCond,
				appName:  defaultIfEmpty(o.AppName, auditAppName),
				severity: defaultIfEmpty(o.Severity, auditSeverity),
				facility: defaultIfEmpty(o.Facility, auditFacility),
			},
		)
	}
}

// buildRFC5424Rules constructs the default rules for RFC5424 format
func buildRFC5424Rules(o *obs.Syslog, rules *[]defaultRule) {
	if o.MsgId == "" {
		*rules = append(*rules, defaultRule{
			msgId: msgId,
		})
	}

	if o.ProcId == "" || o.AppName == "" {
		*rules = append(*rules, defaultRule{
			cond:    isInfrastructureNodeLogCond,
			appName: defaultIfEmpty(o.AppName, nodeAppNameRFC5424),
			procId:  defaultIfEmpty(o.ProcId, nodeProcIdRFC5424),
		})
	}

	if o.AppName == "" || o.ProcId == "" || o.Severity == "" || o.Facility == "" {
		*rules = append(*rules,
			defaultRule{
				cond:     isContainerLogCond,
				appName:  defaultIfEmpty(o.AppName, containerAppNameRFC5424),
				procId:   defaultIfEmpty(o.ProcId, containerProcId),
				severity: defaultIfEmpty(o.Severity, containerSeverity),
				facility: defaultIfEmpty(o.Facility, containerFacility),
			},
			defaultRule{
				cond:     isAuditLogCond,
				appName:  defaultIfEmpty(o.AppName, auditAppName),
				procId:   defaultIfEmpty(o.ProcId, auditProcId),
				severity: defaultIfEmpty(o.Severity, auditSeverity),
				facility: defaultIfEmpty(o.Facility, auditFacility),
			},
		)
	}
}
