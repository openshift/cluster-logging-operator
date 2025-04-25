package syslog

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"net/url"
	"regexp"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
)

const (
	TCP         = `tcp`
	TLS         = `tls`
	ParsedMsg   = "parsed_msg"
	defFacility = `{._internal.syslog.facility || "user"}`
	defSeverity = `{._internal.syslog.severity || "informational"}`
	defProcId   = `{._internal.syslog.proc_id || "-"}`
	defTag      = `{._internal.syslog.tag || ""}`
	defAppName  = `{._internal.syslog.app_name || "-"}`
	defMsgId    = `{._internal.syslog.msg_id || "-"}`

	nodeTag    = `._internal.syslog.tag = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "")`
	nodeProcId = `._internal.syslog.proc_id = to_string!(.systemd.t.PID || "")`

	containerTag = `._internal.syslog.tag = join!([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "")
#Remove non-alphanumeric characters
._internal.syslog.tag = replace(._internal.syslog.tag, r'[^a-zA-Z0-9]', "")
#Truncate the sanitized tag to 32 characters
._internal.syslog.tag = truncate(._internal.syslog.tag, 32)
`
	containerSeverity  = `._internal.syslog.severity = .level`
	containerFacility  = `._internal.syslog.facility = "user"`
	auditTag           = `._internal.syslog.tag = .log_source`
	auditSeverity      = `._internal.syslog.severity = "informational"`
	auditFacility      = `._internal.syslog.facility = "security"`
	msgId              = `._internal.syslog.msg_id = .log_source`
	nodeAppName        = `._internal.syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER||"-")`
	nodeProcId2        = `._internal.syslog.proc_id = to_string!(.systemd.t.PID||"-")`
	containerAppName   = `._internal.syslog.app_name = join!([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "_")`
	containerProcId    = `._internal.syslog.proc_id = to_string!(.kubernetes.pod_id||"-")`
	containerSeverity2 = `._internal.syslog.severity = .level`
	containerFacility2 = `._internal.syslog.facility = "user"`
	auditAppName       = `._internal.syslog.app_name = .log_source`
	auditProcId2       = `._internal.syslog.proc_id = to_string!(.auditID || "-")`

	auditSeverity2 = `._internal.syslog.severity = "informational"`
	auditFacility2 = `._internal.syslog.facility = "security"`
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
{{end}}`
}

type FieldVRLStringPair struct {
	Field     string
	VRLString string
}

type EncodingTemplateField struct {
	FieldVRLList []FieldVRLStringPair
}

type SyslogEncodingRemap struct {
	ComponentID    string
	Inputs         string
	EncodingFields EncodingTemplateField
	PayloadKey     string
	RFC            string
	Defaults       string
	ParseMsg       bool
}

func (ser SyslogEncodingRemap) Name() string {
	return "syslogEncodingRemap"
}

func (ser SyslogEncodingRemap) Template() string {
	return fmt.Sprintf(`{{define "`+ser.Name()+`" -}}
[transforms.{{.ComponentID}}]
type = "remap"
inputs = {{.Inputs}}
source = '''
{{if .Defaults}}
{{.Defaults}}
{{end}}
{{if .EncodingFields.FieldVRLList -}}
{{if .ParseMsg}}
_tmp, err = parse_json(string!(.message))
if err != null {
  _tmp = .
  log(err, level: "error")
} else {
  _tmp = merge!(.,_tmp)
}
%s = _tmp
{{end}}
{{range $templatePair := .EncodingFields.FieldVRLList -}}
	.{{$templatePair.Field}} = {{$templatePair.VRLString}}
{{end -}}
{{end}}

{{if eq .RFC "RFC3164" -}}
if exists(.proc_id) && .proc_id != "-" && .proc_id != "" {
 .tag = .tag + "[" + .proc_id  + "]"
}
{{end}}

{{if .PayloadKey -}}
if is_null({{.PayloadKey}}) {
	.payload_key = .
} else {
	.payload_key = {{.PayloadKey}}
}
{{end}}
'''
{{end -}}
`, ParsedMsg)
}

type SyslogEncoding struct {
	ComponentID  string
	RFC          string
	Facility     string
	Severity     string
	AddLogSource genhelper.OptionalPair
	PayloadKey   genhelper.OptionalPair
}

func (se SyslogEncoding) Name() string {
	return "syslogEncoding"
}

func (se SyslogEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = "syslog"
except_fields = ["_internal"]
rfc = "{{.RFC}}"
{{ .AddLogSource }}
{{ .PayloadKey }}
{{end}}`
}

func (s *Syslog) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	parseEncodingID := vectorhelpers.MakeID(id, "parse_encoding")
	templateFieldPairs, needToParseMsg := getEncodingTemplatesAndFields(*o.Syslog)
	u, _ := url.Parse(o.Syslog.URL)
	sink := Output(id, o, []string{parseEncodingID}, secrets, op, u.Scheme, u.Host)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	syslogElements := []Element{
		parseEncoding(parseEncodingID, inputs, templateFieldPairs, needToParseMsg, o.Syslog),
		sink,
	}

	syslogElements = append(syslogElements, Encoding(id, o, templateFieldPairs.FieldVRLList)...)

	return append(syslogElements,
		common.NewAcknowledgments(id, strategy),
		common.NewBuffer(id, strategy),
		tls.New(id, o.TLS, secrets, op, tls.IncludeEnabledOption),
	)
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op Options, urlScheme string, host string) *Syslog {
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
func getEncodingTemplatesAndFields(s obs.Syslog) (EncodingTemplateField, bool) {
	templateFields := EncodingTemplateField{
		FieldVRLList: []FieldVRLStringPair{},
	}

	var needToParseMsg = false

	appendField := func(fieldName string, value string, defaultVal string) {
		if value == "" {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     fieldName,
				VRLString: commontemplate.TransformUserTemplateToVRL(defaultVal),
			})
		} else {
			if commontemplate.PathRegex.MatchString(value) {
				needToParseMsg = true
			}
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     fieldName,
				VRLString: commontemplate.TransformUserTemplateToVRL(value, ParsedMsg),
			})
		}
	}

	appendField("facility", s.Facility, defFacility)
	appendField("severity", s.Severity, defSeverity)
	appendField("proc_id", s.ProcId, defProcId)

	if s.RFC == obs.SyslogRFC3164 {
		appendField("tag", s.AppName, defTag)
	} else {
		appendField("app_name", s.AppName, defAppName)
		appendField("msg_id", s.MsgId, defMsgId)
	}

	return templateFields, needToParseMsg
}

func Encoding(id string, o obs.OutputSpec, templatePairs []FieldVRLStringPair) []Element {
	sysLEncode := SyslogEncoding{
		ComponentID:  id,
		RFC:          strings.ToLower(string(o.Syslog.RFC)),
		Facility:     Facility(o.Syslog),
		Severity:     Severity(o.Syslog),
		AddLogSource: genhelper.NewOptionalPair("add_log_source", o.Syslog.Enrichment == obs.EnrichmentTypeKubernetesMinimal),
		PayloadKey:   genhelper.NewOptionalPair("payload_key", nil),
	}
	if o.Syslog.PayloadKey != "" {
		sysLEncode.PayloadKey.Value = "payload_key"
	}

	encodingFields := []Element{
		sysLEncode,
	}
	// Add fields that have been templated
	for _, pair := range templatePairs {
		encodingFields = append(encodingFields, KV(pair.Field, fmt.Sprintf(`"$$.message.%s"`, pair.Field)))
	}

	return encodingFields
}

func parseEncoding(id string, inputs []string, templatePairs EncodingTemplateField, needToParseMsg bool, o *obs.Syslog) Element {
	return SyslogEncodingRemap{
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		EncodingFields: templatePairs,
		PayloadKey:     PayloadKey(o.PayloadKey),
		RFC:            string(o.RFC),
		Defaults:       buildDefaults(o),
		ParseMsg:       needToParseMsg,
	}
}

func Facility(s *obs.Syslog) string {
	if s == nil || s.Facility == "" {
		return ""
	}
	if IsKeyExpr(s.Facility) {
		return fmt.Sprintf("$%s", s.Facility)
	}
	return s.Facility
}

func Severity(s *obs.Syslog) string {
	if s == nil || s.Severity == "" {
		return ""
	}
	if IsKeyExpr(s.Severity) {
		return fmt.Sprintf("$%s", s.Severity)
	}
	return s.Severity
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

// The Syslog output fields can be set to an expression of the form $.abc.xyz
// If an expression is used, its value will be taken from corresponding key in the record
// Example: $.message.procid_key
var keyre = regexp.MustCompile(`^\$(\.[[:word:]]*)+$`)

func IsKeyExpr(str string) bool {
	return keyre.MatchString(str)
}

func buildDefaults(o *obs.Syslog) string {
	var builder strings.Builder
	type defaultRule struct {
		cond     string
		appName  string
		procId   string
		severity string
		facility string
	}
	var defaultRules []defaultRule

	if o.RFC == obs.SyslogRFC3164 {
		if o.ProcId == "" || o.AppName == "" {
			defaultRules = append(defaultRules, defaultRule{
				cond:    `.log_type == "infrastructure" && .log_source == "node"`,
				appName: ifEmpty(o.AppName, nodeTag),
				procId:  ifEmpty(o.ProcId, nodeProcId),
			})
		}
		if o.AppName == "" || o.Severity == "" || o.Facility == "" {
			defaultRules = append(defaultRules, defaultRule{
				cond:     `.log_source == "container"`,
				appName:  ifEmpty(o.AppName, containerTag),
				severity: ifEmpty(o.Severity, containerSeverity),
				facility: ifEmpty(o.Facility, containerFacility),
			}, defaultRule{
				cond:     `.log_type == "audit"`,
				appName:  ifEmpty(o.AppName, auditTag),
				severity: ifEmpty(o.Severity, auditSeverity),
				facility: ifEmpty(o.Facility, auditFacility),
			})
		}
	}

	if o.RFC == obs.SyslogRFC5424 {
		builder.WriteString(msgId + "\n")
		if o.ProcId == "" || o.AppName == "" {
			defaultRules = append(defaultRules, defaultRule{
				cond:    `.log_type == "infrastructure" && .log_source == "node"`,
				appName: ifEmpty(o.AppName, nodeAppName),
				procId:  ifEmpty(o.ProcId, nodeProcId2),
			})
		}
		if o.AppName == "" || o.ProcId == "" || o.Severity == "" || o.Facility == "" {
			defaultRules = append(defaultRules, defaultRule{
				cond:     `.log_source == "container"`,
				appName:  ifEmpty(o.AppName, containerAppName),
				procId:   ifEmpty(o.ProcId, containerProcId),
				severity: ifEmpty(o.Severity, containerSeverity),
				facility: ifEmpty(o.Facility, containerFacility),
			}, defaultRule{
				cond:     `.log_type == "audit"`,
				appName:  ifEmpty(o.AppName, auditAppName),
				procId:   ifEmpty(o.ProcId, auditProcId2),
				severity: ifEmpty(o.Severity, auditSeverity2),
				facility: ifEmpty(o.Facility, auditFacility2),
			})
		}
	}

	for _, b := range defaultRules {
		builder.WriteString(fmt.Sprintf("if %s {\n", b.cond))
		if b.appName != "" {
			builder.WriteString(b.appName + "\n")
		}
		if b.procId != "" {
			builder.WriteString(b.procId + "\n")
		}
		if b.severity != "" {
			builder.WriteString(b.severity + "\n")
		}
		if b.facility != "" {
			builder.WriteString(b.facility + "\n")
		}
		builder.WriteString("}\n")
	}

	return builder.String()
}

func ifEmpty(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return ""
}
