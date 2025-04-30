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
	defFacility = `to_string!(._internal.syslog.facility || "user")`
	defSeverity = `to_string!(._internal.syslog.severity || "informational")`
	defProcId   = `to_string!(._internal.syslog.proc_id || "-")`
	defTag      = `to_string!(._internal.syslog.tag || "")`
	defAppName  = `to_string!(._internal.syslog.app_name || "-")`
	defMsgId    = `to_string!(._internal.syslog.msg_id || "-")`

	nodeAppName       = `._internal.syslog.app_name = to_string!(._internal.systemd.u.SYSLOG_IDENTIFIER || "-")`
	nodeTag           = `._internal.syslog.tag = to_string!(._internal.systemd.u.SYSLOG_IDENTIFIER || "")`
	nodeProcIdRFC3164 = `._internal.syslog.proc_id = to_string!(._internal.systemd.t.PID || "")`
	nodeProcIdRFC5424 = `._internal.syslog.proc_id = to_string!(._internal.systemd.t.PID || "-")`

	containerAppName  = `._internal.syslog.app_name = join!([._internal.kubernetes.namespace_name, ._internal.kubernetes.pod_name, ._internal.kubernetes.container_name], "_")`
	containerFacility = `._internal.syslog.facility = "user"`
	containerSeverity = `._internal.syslog.severity = ._internal.level`
	containerProcId   = `._internal.syslog.proc_id = to_string!(._internal.kubernetes.pod_id || "-")`
	containerTag      = `._internal.syslog.tag = join!([._internal.kubernetes.namespace_name, ._internal.kubernetes.pod_name, ._internal.kubernetes.container_name], "")
#Remove non-alphanumeric characters
._internal.syslog.tag = replace(._internal.syslog.tag, r'[^a-zA-Z0-9]', "")
#Truncate the sanitized tag to 32 characters
._internal.syslog.tag = truncate(._internal.syslog.tag, 32)
`

	auditTag      = `._internal.syslog.tag = ._internal.log_source`
	auditSeverity = `._internal.syslog.severity = "informational"`
	auditFacility = `._internal.syslog.facility = "security"`
	auditAppName  = `._internal.syslog.app_name = ._internal.log_source`
	auditProcId   = `._internal.syslog.proc_id = to_string!(._internal.auditID || "-")`

	msgId = `._internal.syslog.msg_id = ._internal.log_source`

	isInfrastructureNodeLogCond = `._internal.log_type == "infrastructure" && ._internal.log_source == "node"`
	isContainerLogCond          = `._internal.log_source == "container"`
	isAuditLogCond              = `._internal.log_type == "audit"`
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
}

func (ser SyslogEncodingRemap) Name() string {
	return "syslogEncodingRemap"
}

func (ser SyslogEncodingRemap) Template() string {
	return `{{define "` + ser.Name() + `" -}}
[transforms.{{.ComponentID}}]
type = "remap"
inputs = {{.Inputs}}
source = '''
{{if .Defaults}}
{{.Defaults}}
{{end}}
{{if .EncodingFields.FieldVRLList -}}
{{range $templatePair := .EncodingFields.FieldVRLList -}}
	.{{$templatePair.Field}} = {{$templatePair.VRLString}}
{{end -}}
{{end}}

{{if eq .RFC "RFC3164" -}}
if exists(.proc_id) && .proc_id != "-" && .proc_id != "" {
 .tag = .tag + "[" + .proc_id + "]"
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
`
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
	templateFieldPairs := getEncodingTemplatesAndFields(*o.Syslog)
	u, _ := url.Parse(o.Syslog.URL)
	sink := Output(id, o, []string{parseEncodingID}, secrets, op, u.Scheme, u.Host)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	syslogElements := []Element{
		parseEncoding(parseEncodingID, inputs, templateFieldPairs, o.Syslog),
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
func getEncodingTemplatesAndFields(s obs.Syslog) EncodingTemplateField {
	templateFields := EncodingTemplateField{
		FieldVRLList: []FieldVRLStringPair{},
	}

	appendField := func(fieldName string, value string, defaultVal string) {
		if value == "" {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     fieldName,
				VRLString: defaultVal,
			})
		} else {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     fieldName,
				VRLString: commontemplate.TransformUserTemplateToVRL(value),
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

	return templateFields
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

func parseEncoding(id string, inputs []string, templatePairs EncodingTemplateField, o *obs.Syslog) Element {
	return SyslogEncodingRemap{
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		EncodingFields: templatePairs,
		PayloadKey:     PayloadKey(o.PayloadKey),
		RFC:            string(o.RFC),
		Defaults:       buildDefaults(o),
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
			appName: emptyToValue(o.AppName, nodeTag),
			procId:  emptyToValue(o.ProcId, nodeProcIdRFC3164),
		})
	}

	if o.AppName == "" || o.Severity == "" || o.Facility == "" {
		*rules = append(*rules,
			defaultRule{
				cond:     isContainerLogCond,
				appName:  emptyToValue(o.AppName, containerTag),
				severity: emptyToValue(o.Severity, containerSeverity),
				facility: emptyToValue(o.Facility, containerFacility),
			},
			defaultRule{
				cond:     isAuditLogCond,
				appName:  emptyToValue(o.AppName, auditTag),
				severity: emptyToValue(o.Severity, auditSeverity),
				facility: emptyToValue(o.Facility, auditFacility),
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
			appName: emptyToValue(o.AppName, nodeAppName),
			procId:  emptyToValue(o.ProcId, nodeProcIdRFC5424),
		})
	}

	if o.AppName == "" || o.ProcId == "" || o.Severity == "" || o.Facility == "" {
		*rules = append(*rules,
			defaultRule{
				cond:     isContainerLogCond,
				appName:  emptyToValue(o.AppName, containerAppName),
				procId:   emptyToValue(o.ProcId, containerProcId),
				severity: emptyToValue(o.Severity, containerSeverity),
				facility: emptyToValue(o.Facility, containerFacility),
			},
			defaultRule{
				cond:     isAuditLogCond,
				appName:  emptyToValue(o.AppName, auditAppName),
				procId:   emptyToValue(o.ProcId, auditProcId),
				severity: emptyToValue(o.Severity, auditSeverity),
				facility: emptyToValue(o.Facility, auditFacility),
			},
		)
	}
}

// writeRulesToBuilder converts the rules collection into formatted string
func writeRulesToBuilder(builder *strings.Builder, rules []defaultRule) {
	for _, rule := range rules {

		if rule.cond != "" {
			builder.WriteString(fmt.Sprintf("if %s {\n", rule.cond))
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

func emptyToValue(val, defaultVal string) string {
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
