package syslog

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
)

const (
	TCP = `tcp`
	TLS = `tls`
)

// getEncodingTemplatesAndFields determines which encoding fields are templated
// so that the templates can be parsed to appropriate VRL
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
	Tag            string
	ProcID         string
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
. = merge(., parse_json!(string!(.message))) ?? .
{{- if eq .RFC "RFC3164" }}
if .log_type == "infrastructure" && ._internal.log_source == "node" {
    ._internal.syslog.tag = to_string!(.systemd.u.SYSLOG_IDENTIFIER || "")
	._internal.syslog.proc_id = to_string!(.systemd.t.PID || "")
}
if ._internal.log_source == "container" {
   	._internal.syslog.tag = join!([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "")
   	._internal.syslog.severity = .level
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
{{- if eq .RFC "RFC5424" }}
._internal.syslog.msg_id = ._internal.log_source
if .log_type == "infrastructure" && ._internal.log_source == "node" {
	._internal.syslog.app_name = to_string!(.systemd.u.SYSLOG_IDENTIFIER||"-")
	._internal.syslog.proc_id = to_string!(.systemd.t.PID||"-")
}
if ._internal.log_source == "container" {
   ._internal.syslog.app_name = join!([.kubernetes.namespace_name, .kubernetes.pod_name, .kubernetes.container_name], "_")
   ._internal.syslog.proc_id = to_string!(.kubernetes.pod_id||"-")
   ._internal.syslog.severity = .level
   ._internal.syslog.facility = "user"
}
if .log_type == "audit" {
   ._internal.syslog.app_name = ._internal.log_source
   ._internal.syslog.proc_id = to_string!(.auditID || "-")
   ._internal.syslog.severity = "informational"
   ._internal.syslog.facility = "security"
}
{{end}}

{{if .EncodingFields.FieldVRLList -}}
{{range $templatePair := .EncodingFields.FieldVRLList -}}
	.{{$templatePair.Field}} = {{$templatePair.VRLString}}
{{end -}}
{{end}}

{{if and (eq .RFC "RFC3164") (eq .Tag "") (eq .ProcID "") -}}
if exists(.proc_id) && .proc_id != "-" && .proc_id != "" {
 .tag = .tag + "[" + .proc_id  + "]"
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
	AppName      string
	ProcID       string
	MsgID        string
	Tag          string
	AddLogSource Element
	PayloadKey   Element
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

{{- if .Facility }}
facility = "{{.Facility}}"
{{- end}}
{{- if .Severity }}
severity = "{{.Severity}}"
{{- end }}
{{- if .ProcID }}
proc_id = "{{.ProcID}}"
{{- end }}
{{- if .AppName }}
app_name = "{{.AppName}}"
{{- end }}
{{- if .MsgID }}
msg_id = "{{.MsgID}}"
{{- end }}
{{- if .Tag }}
tag = "{{.Tag}}"
{{- end }}
{{ optional .AddLogSource -}}
{{ optional .PayloadKey -}}
{{- end }}`
}

func (s *Syslog) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	o.Syslog.RFC = RFC(o.Syslog)
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	parseEncodingID := vectorhelpers.MakeID(id, "parse_encoding")
	templateFieldPairs := getEncodingTemplatesAndFields(o.Syslog)
	u, _ := url.Parse(o.URL)
	sink := Output(id, o, []string{parseEncodingID}, secret, op, u.Scheme, u.Host)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	syslogElements := []Element{
		parseEncoding(parseEncodingID, inputs, templateFieldPairs, o.Syslog),
		sink,
	}

	syslogElements = append(syslogElements, Encoding(id, o, templateFieldPairs.FieldVRLList)...)

	syslogElements = append(syslogElements,
		common.NewAcknowledgments(id, strategy),
		common.NewBuffer(id, strategy))
	return append(syslogElements, TLSConf(id, o, secret, op)...)
}
func getEncodingTemplatesAndFields(s *logging.Syslog) EncodingTemplateField {
	templateFields := EncodingTemplateField{
		FieldVRLList: []FieldVRLStringPair{},
	}

	if s.Facility == "" {
		templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
			Field:     "facility",
			VRLString: vectorhelpers.TransformUserTemplateToVRL(`{._internal.syslog.facility || "user"}`),
		})
	}

	if s.Severity == "" {
		templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
			Field:     "severity",
			VRLString: vectorhelpers.TransformUserTemplateToVRL(`{._internal.syslog.severity || "informational"}`),
		})
	}

	if s.ProcID == "" {
		templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
			Field:     "proc_id",
			VRLString: vectorhelpers.TransformUserTemplateToVRL(`{._internal.syslog.proc_id || "-"}`),
		})
	}

	if s.RFC == logging.SyslogRFC3164 {
		if s.Tag == "" {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     "tag",
				VRLString: vectorhelpers.TransformUserTemplateToVRL(`{._internal.syslog.tag || ""}`),
			})
		}

	} else {
		if s.AppName == "" {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     "app_name",
				VRLString: vectorhelpers.TransformUserTemplateToVRL(`{._internal.syslog.app_name || "-"}`),
			})
		}
		if s.MsgID == "" {
			templateFields.FieldVRLList = append(templateFields.FieldVRLList, FieldVRLStringPair{
				Field:     "msg_id",
				VRLString: vectorhelpers.TransformUserTemplateToVRL(`{._internal.syslog.msg_id || "-"}`),
			})
		}
	}

	return templateFields
}

func Encoding(id string, o logging.OutputSpec, templatePairs []FieldVRLStringPair) []Element {
	sysLEncode := SyslogEncoding{
		ComponentID:  id,
		RFC:          strings.ToLower(o.Syslog.RFC),
		Facility:     Facility(o.Syslog),
		Severity:     Severity(o.Syslog),
		AddLogSource: AddLogSource(o.Syslog),
	}
	if o.Syslog.PayloadKey != "" {
		sysLEncode.PayloadKey = PayloadKey(o.Syslog)
	}

	if o.Syslog.RFC == logging.SyslogRFC5424 {
		sysLEncode.AppName = AppName(o.Syslog)
		sysLEncode.MsgID = MsgID(o.Syslog)
	}
	sysLEncode.ProcID = ProcID(o.Syslog)

	if o.Syslog.RFC == logging.SyslogRFC3164 {
		sysLEncode.Tag = Tag(o.Syslog)
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

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options, urlScheme string, host string) *Syslog {
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

func parseEncoding(id string, inputs []string, templatePairs EncodingTemplateField, o *logging.Syslog) Element {
	return SyslogEncodingRemap{
		ComponentID:    id,
		Inputs:         vectorhelpers.MakeInputs(inputs...),
		EncodingFields: templatePairs,
		RFC:            o.RFC,
		Tag:            o.Tag,
		ProcID:         o.ProcID,
		PayloadKey:     o.PayloadKey,
	}
}

func TLSConf(id string, o logging.OutputSpec, secret *corev1.Secret, op Options) []Element {
	if o.Secret != nil || (o.TLS != nil && o.TLS.InsecureSkipVerify) {
		if tlsConf := common.GenerateTLSConfWithID(id, o, secret, op, false); tlsConf != nil {
			return []Element{tlsConf}
		}
	}
	return []Element{}
}

func Facility(s *logging.Syslog) string {
	if s == nil || s.Facility == "" {
		return ""
	}
	if IsKeyExpr(s.Facility) {
		return fmt.Sprintf("$%s", s.Facility)
	}
	return vectorhelpers.EscapeDollarSigns(s.Facility)
}

func Severity(s *logging.Syslog) string {
	if s == nil || s.Severity == "" {
		return ""
	}
	if IsKeyExpr(s.Severity) {
		return fmt.Sprintf("$%s", s.Severity)
	}
	return vectorhelpers.EscapeDollarSigns(s.Severity)
}

func RFC(s *logging.Syslog) string {
	if s == nil || s.RFC == "" {
		return logging.SyslogRFC5424
	}

	rfc := strings.ToUpper(s.RFC)
	switch rfc {
	case logging.SyslogRFC5424, logging.SyslogRFC3164:
		return rfc
	default:
		return "Unknown RFC"
	}
}

func AppName(s *logging.Syslog) string {
	if s == nil || s.AppName == "" {
		return ""
	}
	if IsKeyExpr(s.AppName) {
		return fmt.Sprintf(`$%s`, s.AppName)
	}
	return vectorhelpers.EscapeDollarSigns(s.AppName)
}

func Tag(s *logging.Syslog) string {
	if s == nil || s.Tag == "" {
		return ""
	}

	var tag string
	if IsKeyExpr(s.Tag) {
		tag = fmt.Sprintf(`$%s`, s.Tag)
	} else {
		tag = vectorhelpers.EscapeDollarSigns(s.Tag)
	}

	if s.ProcID != "" {
		var procID string
		if IsKeyExpr(s.ProcID) {
			procID = fmt.Sprintf(`$%s`, s.ProcID)
		} else {
			procID = vectorhelpers.EscapeDollarSigns(s.ProcID)
		}
		tag = fmt.Sprintf(`%s[%s]`, tag, procID)
	}

	return tag
}

func MsgID(s *logging.Syslog) string {
	if s == nil || s.MsgID == "" {
		return ""
	}
	if IsKeyExpr(s.MsgID) {
		return fmt.Sprintf(`$%s`, s.MsgID)
	}
	return vectorhelpers.EscapeDollarSigns(s.MsgID)
}

func ProcID(s *logging.Syslog) string {
	if s == nil || s.ProcID == "" {
		return ""
	}
	if IsKeyExpr(s.ProcID) {
		return fmt.Sprintf(`$%s`, s.ProcID)
	}
	return vectorhelpers.EscapeDollarSigns(s.ProcID)
}

func AddLogSource(s *logging.Syslog) Element {
	if s == nil || !s.AddLogSource {
		return Nil
	}
	return KV("add_log_source", "true")
}

func PayloadKey(s *logging.Syslog) Element {
	if s == nil || s.PayloadKey == "" {
		return Nil
	}
	return KV("payload_key", fmt.Sprintf(`"%s"`, s.PayloadKey))
}

// The Syslog output fields can be set to an expression of the form $.abc.xyz
// If an expression is used, its value will be taken from corresponding key in the record
// Example: $.message.procid_key
var keyre = regexp.MustCompile(`^\$(\.[[:word:]]*)+$`)

func IsKeyExpr(str string) bool {
	return keyre.MatchString(str)
}
