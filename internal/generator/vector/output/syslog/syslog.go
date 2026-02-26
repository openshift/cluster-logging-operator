package syslog

import (
	"fmt"
	"net/url"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
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

	defProcId = `to_string!(._syslog.proc_id || "")
if exists(._syslog.proc_id) && is_empty(strip_whitespace(string!(._syslog.proc_id))) { del(._syslog.proc_id) }
`
	defTag    = `to_string!(._syslog.app_name || "")`

	// Default values for Syslog fields for infrastructure logType if source 'node'
	nodeAppName       = `._syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "-")`
	nodeTag           = `._syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "")`
	nodeProcIdRFC3164 = `._syslog.proc_id = to_string!(.systemd.t.PID || "")`
	nodeProcIdRFC5424 = `._syslog.proc_id = to_string!(.systemd.t.PID || "-")`

	// Default values for Syslog fields for application logType and infrastructure logType if source 'container'
	containerAppName = `._syslog.app_name, err = join([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "_")
if err != null {
  log("K8s metadata (namespace, pod, or container) missing; syslog.app_name set to '-'", level: "error") 
  ._syslog.app_name = "-"
}
`
	containerFacility = `._syslog.facility = "user"`
	containerSeverity = `._syslog.severity = .level`
	containerProcId   = `._syslog.proc_id = to_string!(.kubernetes.pod_id || "")`
	containerTag      = `._syslog.app_name, err = join([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "")
if err != null {
  log("K8s metadata (namespace, pod, or container) missing: unable to calculate syslog.app_name (TAG)", level: "error") 
} else {
  #Remove non-alphanumeric characters
  ._syslog.app_name = replace(._syslog.app_name, r'[^a-zA-Z0-9]', "")
  #Truncate the sanitized tag to 32 characters
  ._syslog.app_name = truncate(._syslog.app_name, 32)
}
`

	// Default values for Syslog fields for audit logType
	auditTag      = `._syslog.app_name = .log_source`
	auditSeverity = `._syslog.severity = "informational"`
	auditFacility = `._syslog.facility = "security"`
	auditAppName  = `._syslog.app_name = .log_source`
	auditProcId   = `._syslog.proc_id = to_string!(.auditID || "-")`

	msgId = `._syslog.msg_id = .log_source`

	// conditions
	isInfrastructureNodeLogCond = `.log_type == "infrastructure" && .log_source == "node"`
	isContainerLogCond          = `.log_source == "container"`
	isAuditLogCond              = `.log_type == "audit"`
)

type Syslog struct {
	ComponentID string
	Inputs      string
	Address     string
	Mode        string
	common.RootMixin
}

func (s Syslog) Name() string {
	return "SyslogVectorTemplate"
}

func (s Syslog) Template() string {
	return `{{define "` + s.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "socket"
inputs = {{.Inputs}}
address = "{{.Address}}"
mode = "{{.Mode}}"
{{if eq .Mode "tcp"}}
keepalive.time_secs = 60
{{end}}
{{end}}`
}

type FieldVRLStringPair struct {
	Field     string
	VRLString string
}

type EncodingTemplateField struct {
	FieldVRLList []FieldVRLStringPair
}

type RemapEncodingFields struct {
	ComponentID    string
	Inputs         string
	EncodingFields EncodingTemplateField
	PayloadKey     string
	RFC            string
	Defaults       string
	EnrichmentType   string
}

func (ser RemapEncodingFields) Name() string {
	return "syslogEncodingRemap"
}

func (ser RemapEncodingFields) Template() string {
	return `{{define "` + ser.Name() + `" -}}
[transforms.{{.ComponentID}}]
type = "remap"
inputs = {{.Inputs}}
source = '''
# set .hostname to the '.host', required by syslog encoder   
hostname = to_string(.hostname) ?? ""
if !is_empty(strip_whitespace(hostname)) {
  .host = .hostname
}
._syslog = {}
{{if .Defaults}}
{{.Defaults}}
{{end}}
{{if .EncodingFields.FieldVRLList -}}
{{range $templatePair := .EncodingFields.FieldVRLList -}}
	{{$templatePair.Field}} = {{$templatePair.VRLString}}
{{end -}}
{{end}}
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
}

{{if .PayloadKey -}}
# Payload key configured, going to use the payload key for .message field 
payload_key = {{.PayloadKey}}
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
}
{{ else }}
# Payload key NOT configured, full payload set to .message field (skipping internal objects)
excluded_fields = ["_internal", "_syslog"]
temp = .
for_each(excluded_fields) -> |_index, field| {
  temp = remove(temp, [field]) ?? temp
}
.message = temp
{{ end }}

{{ if eq .EnrichmentType "KubernetesMinimal" }}
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
           ", message=" + msg_value

{{ end }}
'''
{{end -}}
`
}

type SyslogEncoding struct {
	ComponentID  string
	RFC          string
	Facility     genhelper.OptionalPair
	Severity     genhelper.OptionalPair
	AppName      genhelper.OptionalPair
	ProcID       genhelper.OptionalPair
	Tag          genhelper.OptionalPair
	MsgID        genhelper.OptionalPair
}

func (se SyslogEncoding) Name() string {
	return "syslogEncoding"
}

func (se SyslogEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = "syslog"
except_fields = ["_internal"]
syslog.rfc = "{{.RFC}}"
{{ .Facility }}
{{ .Severity }}
{{ .AppName }}   
{{ .ProcID }}
{{ .MsgID }}
{{end}}`
}

func (s *Syslog) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	parseEncodingID := vectorhelpers.MakeID(id, "parse_encoding")
	templateFieldPairs := getEncodingTemplatesAndFields(*o.Syslog)
	u, _ := url.Parse(o.Syslog.URL)
	sink := Output(id, o, []string{parseEncodingID}, secrets, op, u.Scheme, u.Host)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	syslogElements := []framework.Element{
		parseEncoding(parseEncodingID, inputs, templateFieldPairs, o.Syslog),
		sink,
	}

	syslogElements = append(syslogElements, Encoding(id, o)...)

	return append(syslogElements,
		common.NewAcknowledgments(id, strategy),
		common.NewBuffer(id, strategy),
		tls.New(id, o.TLS, secrets, op, tls.IncludeEnabledOption),
	)
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op utils.Options, urlScheme string, host string) *Syslog {
	var mode = strings.ToLower(urlScheme)
	if urlScheme == TLS {
		mode = TCP
	}
	return &Syslog{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Address:     host,
		Mode:        mode,
		RootMixin:   common.NewRootMixin(nil),
	}
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
		appendField(vrlKeySyslogProcID, s.ProcId, defProcId)
		appendField(vrlKeySyslogAppName, s.AppName, defTag)
	} else {
		appendField(vrlKeySyslogProcID, s.ProcId, "")
		appendField(vrlKeySyslogAppName, s.AppName, "")
		appendField(vrlKeySyslogMsgID, s.MsgId, "")
	}

	return templateFields
}

func Encoding(id string, o obs.OutputSpec) []framework.Element {
	sysLEncode := SyslogEncoding{
		ComponentID:  id,
		RFC:          strings.ToLower(string(o.Syslog.RFC)),
		Facility:     syslogEncodeField("syslog.facility"),
		Severity:     syslogEncodeField("syslog.severity"),
		AppName:      AppName(o.Syslog),
		ProcID:       syslogEncodeField("syslog.proc_id"),
		MsgID:        MsgID(o.Syslog),
		//AddLogSource: genhelper.NewOptionalPair("add_log_source", o.Syslog.Enrichment == obs.EnrichmentTypeKubernetesMinimal),
		//PayloadKey:   genhelper.NewOptionalPair("payload_key", nil),
	}
	//if o.Syslog.PayloadKey != "" {
	//	sysLEncode.PayloadKey.Value = "payload_key"
	//}

	encodingFields := []framework.Element{
		sysLEncode,
	}

	return encodingFields
}

func parseEncoding(id string, inputs []string, templatePairs EncodingTemplateField, o *obs.Syslog) framework.Element {
	return RemapEncodingFields{
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		EncodingFields: templatePairs,
		PayloadKey:     PayloadKey(o.PayloadKey),
		RFC:            string(o.RFC),
		Defaults:       buildDefaults(o),
		EnrichmentType: string(o.Enrichment),
	}
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

func AppName(s *obs.Syslog) genhelper.OptionalPair {
	//if obs.SyslogRFC5424 != s.RFC {
	//	return genhelper.NewNilPair()
	//}
	return syslogEncodeField("syslog.app_name")
}

func MsgID(s *obs.Syslog) genhelper.OptionalPair {
	if obs.SyslogRFC5424 != s.RFC {
		return genhelper.NewNilPair()
	}
	return syslogEncodeField("syslog.msg_id")
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

func syslogEncodeField(field string) genhelper.OptionalPair {
	return genhelper.NewOptionalPair(field, fmt.Sprintf("._%s", field))
}
