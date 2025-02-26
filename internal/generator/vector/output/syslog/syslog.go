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
	TCP = `tcp`
	TLS = `tls`
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
._internal = merge(._internal, parse_json!(string!(._internal.message))) ?? ._internal

{{if eq .RFC "RFC3164" -}}
if ._internal.log_type == "infrastructure" && ._internal.log_source == "node" {
    ._internal.syslog.tag = to_string!(._internal.systemd.u.SYSLOG_IDENTIFIER || "")
	._internal.syslog.proc_id = to_string!(._internal.systemd.t.PID || "")
}
if ._internal.log_source == "container" {
   	._internal.syslog.tag = join!([._internal.kubernetes.namespace_name, ._internal.kubernetes.pod_name, ._internal.kubernetes.container_name], "")
   	._internal.syslog.severity = ._internal.level
   	._internal.syslog.facility = "user"
   	#Remove non-alphanumeric characters
   	._internal.syslog.tag = replace(._internal.syslog.tag, r'[^a-zA-Z0-9]', "")
	#Truncate the sanitized tag to 32 characters
	._internal.syslog.tag = truncate(._internal.syslog.tag, 32)

}
if .log_type == "audit" {
   ._internal.syslog.tag = ._internal.log_source
   ._internal.syslog.severity = "informational"
   ._internal.syslog.facility = "security" 
}
{{end}}

{{if eq .RFC "RFC5424" -}}
._internal.syslog.msg_id = ._internal.log_source

if ._internal.log_type == "infrastructure" && ._internal.log_source == "node" {
	._internal.syslog.app_name = to_string!(._internal.systemd.u.SYSLOG_IDENTIFIER||"-")
	._internal.syslog.proc_id = to_string!(._internal.systemd.t.PID||"-")
}
if ._internal.log_source == "container" {
   ._internal.syslog.app_name = join!([._internal.kubernetes.namespace_name, ._internal.kubernetes.pod_name, ._internal.kubernetes.container_name], "_")
   ._internal.syslog.proc_id = to_string!(._internal.kubernetes.pod_id||"-")
   ._internal.syslog.severity = ._internal.level
   ._internal.syslog.facility = "user"
}
if ._internal.log_type == "audit" {
   ._internal.syslog.app_name = ._internal.log_source
   ._internal.syslog.proc_id = to_string!(._internal.auditID || "-")
   ._internal.syslog.severity = "informational"
   ._internal.syslog.facility = "security"
}
{{end}}

{{if .EncodingFields.FieldVRLList -}}
{{range $templatePair := .EncodingFields.FieldVRLList -}}
	.{{$templatePair.Field}} = {{$templatePair.VRLString}}
{{end -}}
{{end}}

{{if eq .RFC "RFC3164" -}}
if .proc_id != "-" && .proc_id != "" {
  .tag = to_string(.tag||"") + "[" + to_string(.proc_id)  + "]"
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

	if s.Facility == "" {
		s.Facility = `{.syslog.facility || "user"}`
	}
	templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
		Field:     "facility",
		VRLString: commontemplate.TransformUserTemplateToVRL(s.Facility),
	})

	if s.Severity == "" {
		s.Severity = `{.syslog.severity || "informational"}`
	}
	templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
		Field:     "severity",
		VRLString: commontemplate.TransformUserTemplateToVRL(s.Severity),
	})

	if s.ProcId == "" {
		s.ProcId = `{.syslog.proc_id || "-"}`
	}
	templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
		Field:     "proc_id",
		VRLString: commontemplate.TransformUserTemplateToVRL(s.ProcId),
	})

	if s.RFC == obs.SyslogRFC3164 {
		if s.AppName == "" {
			s.AppName = `{.syslog.tag || ""}`
		}
		templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
			Field:     "tag",
			VRLString: commontemplate.TransformUserTemplateToVRL(s.AppName),
		})

	} else {
		if s.AppName == "" {
			s.AppName = `{.syslog.app_name || "-"}`
		}
		templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
			Field:     "app_name",
			VRLString: commontemplate.TransformUserTemplateToVRL(s.AppName),
		})

		if s.MsgId == "" {
			s.MsgId = `{.syslog.msg_id || "-"}`
		}
		templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
			Field:     "msg_id",
			VRLString: commontemplate.TransformUserTemplateToVRL(s.MsgId),
		})
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
