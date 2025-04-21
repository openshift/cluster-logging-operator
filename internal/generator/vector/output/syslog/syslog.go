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
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	corev1 "k8s.io/api/core/v1"
)

const (
	TCP     = `tcp`
	TLS     = `tls`
	RFC3164 = `rfc3164`
	RFC5424 = `rfc5424`
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

type SyslogEncoding struct {
	ComponentID  string
	RFC          string
	Facility     string
	Severity     string
	AppName      Element
	ProcID       Element
	MsgID        Element
	Tag          Element
	AddLogSource Element
	PayloadKey   Element
}

func CopyFieldsFromPayload(id, input string, s *logging.Syslog) (Element, bool) {
	var needToRemap = false
	if s == nil {
		return Nil, false
	}
	var builder strings.Builder
	fields := []string{
		s.Severity,
		s.Facility,
		s.Tag,
		s.AppName,
		s.ProcID,
		s.MsgID,
	}

	payloadKey := "message"
	if s.PayloadKey != "" {
		payloadKey = s.PayloadKey
	}

	tmpl := `_tmp, err = parse_json(string!(.%s))
if err != null { 
	log(err, level: "error") 
} else {
   `
	builder.WriteString(fmt.Sprintf(tmpl, payloadKey))

	for _, field := range fields {
		if IsKeyExpr(field) {
			needToRemap = true
			last := trimFirstSegment(field)
			builder.WriteString(fmt.Sprintf(".%s = _tmp.%s\n", last, last))
		}
	}

	builder.WriteString("}\n")

	if needToRemap {
		return Remap{
			ComponentID: id,
			Inputs:      fmt.Sprintf("[%q]", input),
			VRL:         builder.String(),
		}, needToRemap
	}
	return Nil, needToRemap
}

func trimFirstSegment(keyExpr string) string {
	keyExpr = strings.TrimPrefix(keyExpr, "$.")
	parts := strings.Split(keyExpr, ".")
	if len(parts) <= 1 {
		return keyExpr
	}
	return strings.Join(parts[1:], ".")
}

func (se SyslogEncoding) Name() string {
	return "syslogEncoding"
}

func (se SyslogEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = "syslog"
rfc = "{{.RFC}}"
facility = "{{.Facility}}"
severity = "{{.Severity}}"
{{optional .AppName -}}
{{optional .MsgID -}}
{{optional .ProcID -}}
{{optional .Tag -}}
{{optional .AddLogSource -}}
{{optional .PayloadKey -}}
{{end}}`
}

func (s *Syslog) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	u, _ := url.Parse(o.URL)
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	dedot := normalize.DedotLabels(dedottedID, inputs)

	sinkInput := dedottedID

	jsonID := vectorhelpers.MakeID(id, "json")
	merge, needToRemap := CopyFieldsFromPayload(jsonID, dedottedID, o.Syslog)
	if needToRemap {
		sinkInput = jsonID
	}

	sink := Output(id, o, []string{sinkInput}, secret, op, u.Scheme, u.Host)

	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(
		[]Element{
			dedot,
			merge,
			sink,
			Encoding(id, o),
			common.NewAcknowledgments(id, strategy),
			common.NewBuffer(id, strategy),
		},
		TLSConf(id, o, secret, op),
	)
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

func Encoding(id string, o logging.OutputSpec) Element {
	return SyslogEncoding{
		ComponentID:  id,
		RFC:          RFC(o.Syslog),
		Facility:     Facility(o.Syslog),
		Severity:     Severity(o.Syslog),
		AppName:      AppName(o.Syslog),
		ProcID:       ProcID(o.Syslog),
		MsgID:        MsgID(o.Syslog),
		Tag:          Tag(o.Syslog),
		AddLogSource: AddLogSource(o.Syslog),
		PayloadKey:   PayloadKey(o.Syslog),
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
		return "user"
	}
	if IsKeyExpr(s.Facility) {
		return fmt.Sprintf("$%s", s.Facility)
	}
	return vectorhelpers.EscapeDollarSigns(s.Facility)
}

func Severity(s *logging.Syslog) string {
	if s == nil || s.Severity == "" {
		return "informational"
	}
	if IsKeyExpr(s.Severity) {
		return fmt.Sprintf("$%s", s.Severity)
	}
	return vectorhelpers.EscapeDollarSigns(s.Severity)
}

func RFC(s *logging.Syslog) string {
	if s == nil || s.RFC == "" {
		return RFC5424
	}
	switch strings.ToLower(s.RFC) {
	case RFC3164:
		return RFC3164
	case RFC5424:
		return RFC5424
	}
	return "Unknown RFC"
}

func AppName(s *logging.Syslog) Element {
	if s == nil {
		return Nil
	}
	appname := "app_name"
	if s.AppName == "" {
		return Nil
	}
	if IsKeyExpr(s.AppName) {
		return KV(appname, fmt.Sprintf(`"$%s"`, s.AppName))
	}
	if s.AppName == "tag" {
		return KV(appname, "${tag}")
	}
	return KV(appname, fmt.Sprintf(`"%s"`, vectorhelpers.EscapeDollarSigns(s.AppName)))
}

func Tag(s *logging.Syslog) Element {
	if s == nil || s.Tag == "" {
		return Nil
	}
	tag := "tag"
	if IsKeyExpr(s.Tag) {
		return KV(tag, fmt.Sprintf(`"$%s"`, s.Tag))
	}
	return KV(tag, fmt.Sprintf(`"%s"`, vectorhelpers.EscapeDollarSigns(s.Tag)))
}

func MsgID(s *logging.Syslog) Element {
	if s == nil || s.MsgID == "" {
		return Nil
	}
	msgid := "msg_id"
	if IsKeyExpr(s.MsgID) {
		return KV(msgid, fmt.Sprintf(`"$%s"`, s.MsgID))
	}
	return KV(msgid, fmt.Sprintf(`"%s"`, vectorhelpers.EscapeDollarSigns(s.MsgID)))
}

func ProcID(s *logging.Syslog) Element {
	if s == nil || s.ProcID == "" {
		return Nil
	}
	procid := "proc_id"
	if IsKeyExpr(s.ProcID) {
		return KV(procid, fmt.Sprintf(`"$%s"`, s.ProcID))
	}
	return KV(procid, fmt.Sprintf(`"%s"`, vectorhelpers.EscapeDollarSigns(s.ProcID)))
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
