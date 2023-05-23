package syslog

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
)

const (
	SyslogComponentType = `syslog`
	TCP                 = `tcp`
	TLS                 = `tls`
	RFC3164             = `rfc3164`
	RFC5424             = `rfc5424`
)

type Syslog struct {
	ComponentID string
	Inputs      string
	Address     string
	Mode        string
}

func (s Syslog) Name() string {
	return "SyslogVectorTemplate"
}

func (s Syslog) Template() string {
	return `{{define "` + s.Name() + `" -}}
[transforms.{{.ComponentID}}_json]
type = "remap"
inputs = {{.Inputs}}
source = '''
. = merge(., parse_json!(string!(.message))) ?? .
'''

[sinks.{{.ComponentID}}]
type = "socket"
inputs = ["{{.ComponentID}}_json"]
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

func (se SyslogEncoding) Name() string {
	return "syslogEncoding"
}

func (se SyslogEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding.codec.syslog]
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

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(vectorhelpers.FormatComponentID(o.Name), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	u, _ := url.Parse(o.URL)
	return MergeElements(
		[]Element{
			Output(o, inputs, secret, op, u.Scheme, u.Host),
			Encoding(o),
		},
		TLSConf(o, secret, op),
	)
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options, urlScheme string, host string) Element {
	var mode = strings.ToLower(urlScheme)
	if urlScheme == TLS {
		mode = TCP
	}
	return Syslog{
		ComponentID: vectorhelpers.FormatComponentID(o.Name),
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Address:     host,
		Mode:        mode,
	}
}

func Encoding(o logging.OutputSpec) Element {
	return SyslogEncoding{
		ComponentID:  vectorhelpers.FormatComponentID(o.Name),
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

func TLSConf(o logging.OutputSpec, secret *corev1.Secret, op Options) []Element {
	if o.Secret != nil {
		if tlsConf := security.GenerateTLSConf(o, secret, op, false); tlsConf != nil {
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
	return s.Facility
}

func Severity(s *logging.Syslog) string {
	if s == nil || s.Severity == "" {
		return "informational"
	}
	if IsKeyExpr(s.Severity) {
		return fmt.Sprintf("$%s", s.Severity)
	}
	return s.Severity
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
	return KV(appname, fmt.Sprintf(`"%s"`, s.AppName))
}

func Tag(s *logging.Syslog) Element {
	if s == nil || s.Tag == "" {
		return Nil
	}
	tag := "tag"
	if IsKeyExpr(s.Tag) {
		return KV(tag, fmt.Sprintf(`"$%s"`, s.Tag))
	}
	return KV(tag, fmt.Sprintf(`"%s"`, s.Tag))
}

func MsgID(s *logging.Syslog) Element {
	if s == nil || s.MsgID == "" {
		return Nil
	}
	msgid := "msg_id"
	if IsKeyExpr(s.MsgID) {
		return KV(msgid, fmt.Sprintf(`"$%s"`, s.MsgID))
	}
	return KV(msgid, fmt.Sprintf(`"%s"`, s.MsgID))
}

func ProcID(s *logging.Syslog) Element {
	if s == nil || s.ProcID == "" {
		return Nil
	}
	procid := "proc_id"
	if IsKeyExpr(s.ProcID) {
		return KV(procid, fmt.Sprintf(`"$%s"`, s.ProcID))
	}
	return KV(procid, fmt.Sprintf(`"%s"`, s.ProcID))
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
