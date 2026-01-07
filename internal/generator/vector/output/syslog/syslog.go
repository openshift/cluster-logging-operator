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
	vrlKeyRFC3164ProcID  = "proc_id" // for RFC3164 proc_id
	vrlKeyRFC3164Tag     = "tag"     //for RFC3164 tag
	vrlKeySyslogProcID   = "._syslog.proc_id"
	vrlKeySyslogAppName  = "._syslog.app_name"
	vrlKeySyslogMsgID    = "._syslog.msg_id"

	defProcId = `to_string!(._syslog.proc_id || "-")`
	defTag    = `to_string!(._syslog.tag || "")`

	// Default values for Syslog fields for infrastructure logType if source 'node'
	nodeAppName       = `._syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "-")`
	nodeTag           = `._syslog.tag = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "")`
	nodeProcIdRFC3164 = `._syslog.proc_id = to_string!(.systemd.t.PID || "")`
	nodeProcIdRFC5424 = `._syslog.proc_id = to_string!(.systemd.t.PID || "-")`

	// Default values for Syslog fields for application logType and infrastructure logType if source 'container'
	containerAppName = `._syslog.app_name, err = join([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "_")
if err != null {
  log("K8s metadata (namespace, pod, or container) missing; syslog.appname set to '-'", level: "error") 
  ._syslog.app_name = "-"
}
`
	containerFacility = `._syslog.facility = "user"`
	containerSeverity = `._syslog.severity = .level`
	containerProcId   = `._syslog.proc_id = to_string!(.kubernetes.pod_id || "-")`
	containerTag      = `._syslog.tag, err = join([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "")
if err != null {
  log("K8s metadata (namespace, pod, or container) missing; syslog.tag set to empty", level: "error") 
  ._syslog.tag = ""
} else {
  #Remove non-alphanumeric characters
  ._syslog.tag = replace(._syslog.tag, r'[^a-zA-Z0-9]', "")
  #Truncate the sanitized tag to 32 characters
  ._syslog.tag = truncate(._syslog.tag, 32)
}
`

	// Default values for Syslog fields for audit logType
	auditTag      = `._syslog.tag = .log_source`
	auditSeverity = `._syslog.severity = "informational"`
	auditFacility = `._syslog.facility = "security"`
	auditAppName  = `._syslog.app_name = .log_source`
	auditProcId   = `._syslog.proc_id = to_string!(.auditID || "-")`

	msgId = `._syslog.msg_id = .log_source`

	// conditions
	isInfrastructureNodeLogCond = `.log_type == "infrastructure" && .log_source == "node"`
	isContainerLogCond          = `.log_source == "container"`
	isAuditLogCond              = `.log_type == "audit"`
	defaultPayloadKey           = "payload_key"
)

type FieldVRLStringPair struct {
	Field     string
	VRLString string
}

type EncodingTemplateField struct {
	FieldVRLList []FieldVRLStringPair
}

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	parseEncodingID := vectorhelpers.MakeID(id, "parse_encoding")
	templateFieldPairs := getEncodingTemplatesAndFields(*o.Syslog)
	tfs = api.Transforms{}
	tfs[parseEncodingID] = parseEncoding(inputs, templateFieldPairs, o.Syslog)

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
	scheme := strings.ToLower(urlScheme)
	switch scheme {
	case TLS:
		return sinks.SocketModeTCP
	case TCP:
		return sinks.SocketModeTCP
	case "udp":
		return sinks.SocketModeUDP
	default:
		// Pass through unknown schemes for vector validation
		return sinks.SocketMode(scheme)
	}
}

func buildSocketEncoding(o obs.OutputSpec) *sinks.SocketEncoding {
	encoding := &sinks.SocketEncoding{
		Codec:        "syslog",
		ExceptFields: []string{"_internal"},
		RFC:          strings.ToLower(string(o.Syslog.RFC)),
		Facility:     "$$._syslog.facility",
		Severity:     "$$._syslog.severity",
		ProcID:       "$$._syslog.proc_id",
	}

	// RFC-specific fields
	switch o.Syslog.RFC {
	case obs.SyslogRFC5424:
		encoding.AppName = "$$._syslog.app_name"
		encoding.MsgID = "$$._syslog.msg_id"
	case obs.SyslogRFC3164:
		encoding.Tag = "$$._syslog.tag"
	}

	// Add log source
	if o.Syslog.Enrichment == obs.EnrichmentTypeKubernetesMinimal {
		addLogSource := true
		encoding.AddLogSource = &addLogSource
	} else {
		addLogSource := false
		encoding.AddLogSource = &addLogSource
	}

	// Payload key
	if o.Syslog.PayloadKey != "" {
		encoding.PayloadKey = defaultPayloadKey
	}

	return encoding
}

// getEncodingTemplatesAndFields determines which encoding fields are templated
// so that the templates can be parsed to appropriate VRL
func getEncodingTemplatesAndFields(s obs.Syslog) EncodingTemplateField {
	templateFields := EncodingTemplateField{
		FieldVRLList: []FieldVRLStringPair{},
	}

	appendField := func(fieldName string, userTemplateValue string, defaultValue string) {
		var vrlString string

		if userTemplateValue != "" {
			vrlString = commontemplate.TransformUserTemplateToVRL(userTemplateValue)
		} else {
			vrlString = defaultValue
		}

		if vrlString != "" {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     fieldName,
				VRLString: vrlString,
			})
		}
	}

	appendField(vrlKeySyslogFacility, s.Facility, "")
	appendField(vrlKeySyslogSeverity, s.Severity, "")

	if s.RFC == obs.SyslogRFC3164 {
		appendField(vrlKeyRFC3164ProcID, s.ProcId, defProcId)
		appendField(vrlKeyRFC3164Tag, s.AppName, defTag)
	} else {
		appendField(vrlKeySyslogProcID, s.ProcId, "")
		appendField(vrlKeySyslogAppName, s.AppName, "")
		appendField(vrlKeySyslogMsgID, s.MsgId, "")
	}

	return templateFields
}

func parseEncoding(inputs []string, templatePairs EncodingTemplateField, o *obs.Syslog) types.Transform {
	vrls := []string{buildDefaults(o)}
	for _, tf := range templatePairs.FieldVRLList {
		vrls = append(vrls, fmt.Sprintf("%s = %s", tf.Field, tf.VRLString))
	}
	if o.RFC == obs.SyslogRFC3164 {
		vrls = append(vrls, `
if proc_id != "-" && proc_id != "" {
  ._syslog.tag = to_string(tag||"") + "[" + to_string(proc_id)  + "]"
} else {
  ._syslog.tag = to_string(tag)
}
`)
	}
	if key := PayloadKey(o.PayloadKey); key != "" {
		vrls = append(vrls, fmt.Sprintf(`
if is_null(%s) {
	.payload_key = .
} else {
	.payload_key = %s
}
`, key, key))
	}
	return transforms.NewRemap(strings.Join(vrls, "\n"), inputs...)
}

// PayloadKey returns the whole message or if user templated, uses the specified field from the message.
// This defaults to the whole message
func PayloadKey(plKey string) string {
	// Default
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
	var defaultRules []defaultRule

	switch o.RFC {
	case obs.SyslogRFC3164:
		buildRFC3164Rules(o, &defaultRules)
	case obs.SyslogRFC5424:
		buildRFC5424Rules(o, &defaultRules)
	default:
	}

	writeRulesToBuilder(&builder, defaultRules)

	return builder.String()
}

// buildRFC3164Rules constructs the default rules for RFC3164 format
func buildRFC3164Rules(o *obs.Syslog, rules *[]defaultRule) {
	if o.ProcId == "" || o.AppName == "" {
		*rules = append(*rules, defaultRule{
			cond:    isInfrastructureNodeLogCond,
			appName: getDefaultForEmpty(o.AppName, nodeTag),
			procId:  getDefaultForEmpty(o.ProcId, nodeProcIdRFC3164),
		})
	}

	if o.AppName == "" || o.Severity == "" || o.Facility == "" {
		*rules = append(*rules,
			defaultRule{
				cond:     isContainerLogCond,
				appName:  getDefaultForEmpty(o.AppName, containerTag),
				severity: getDefaultForEmpty(o.Severity, containerSeverity),
				facility: getDefaultForEmpty(o.Facility, containerFacility),
			},
			defaultRule{
				cond:     isAuditLogCond,
				appName:  getDefaultForEmpty(o.AppName, auditTag),
				severity: getDefaultForEmpty(o.Severity, auditSeverity),
				facility: getDefaultForEmpty(o.Facility, auditFacility),
			},
		)
	}
}

// buildRFC5424Rules constructs the default rules for RFC5424 format
func buildRFC5424Rules(o *obs.Syslog, rules *[]defaultRule) {
	if o.MsgId == "" {
		*rules = append(*rules, defaultRule{
			cond:  "",
			msgId: msgId,
		})
	}

	if o.ProcId == "" || o.AppName == "" {
		*rules = append(*rules, defaultRule{
			cond:    isInfrastructureNodeLogCond,
			appName: getDefaultForEmpty(o.AppName, nodeAppName),
			procId:  getDefaultForEmpty(o.ProcId, nodeProcIdRFC5424),
		})
	}

	if o.AppName == "" || o.ProcId == "" || o.Severity == "" || o.Facility == "" {
		*rules = append(*rules,
			defaultRule{
				cond:     isContainerLogCond,
				appName:  getDefaultForEmpty(o.AppName, containerAppName),
				procId:   getDefaultForEmpty(o.ProcId, containerProcId),
				severity: getDefaultForEmpty(o.Severity, containerSeverity),
				facility: getDefaultForEmpty(o.Facility, containerFacility),
			},
			defaultRule{
				cond:     isAuditLogCond,
				appName:  getDefaultForEmpty(o.AppName, auditAppName),
				procId:   getDefaultForEmpty(o.ProcId, auditProcId),
				severity: getDefaultForEmpty(o.Severity, auditSeverity),
				facility: getDefaultForEmpty(o.Facility, auditFacility),
			},
		)
	}
}

// writeRulesToBuilder converts the rules collection into formatted string
func writeRulesToBuilder(builder *strings.Builder, rules []defaultRule) {
	for _, rule := range rules {

		if rule.cond != "" {
			fmt.Fprintf(builder, "if %s {\n", rule.cond)
		}

		writeIfNotEmpty(builder, rule.appName)
		writeIfNotEmpty(builder, rule.procId)
		writeIfNotEmpty(builder, rule.severity)
		writeIfNotEmpty(builder, rule.facility)
		writeIfNotEmpty(builder, rule.msgId)

		if rule.cond != "" {
			builder.WriteString("}\n")
		}
	}
}

func getDefaultForEmpty(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return ""
}

// writeIfNotEmpty adds a string followed by newline to the builder if string is not empty
func writeIfNotEmpty(builder *strings.Builder, s string) {
	if s != "" {
		builder.WriteString(s + "\n")
	}
}
