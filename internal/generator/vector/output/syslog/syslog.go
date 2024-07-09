package syslog

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"net/url"
	"regexp"
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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
[sinks.{{.ComponentID}}.encoding]
codec = "syslog"
except_fields = ["_internal"]
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

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	u, _ := url.Parse(o.Syslog.URL)
	sink := Output(id, o, inputs, secrets, op, u.Scheme, u.Host)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return []Element{
		sink,
		Encoding(id, o),
		common.NewAcknowledgments(id, strategy),
		common.NewBuffer(id, strategy),
		tls.New(id, o.TLS, secrets, op, tls.IncludeEnabledOption),
	}
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options, urlScheme string, host string) *Syslog {
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

func Encoding(id string, o obs.OutputSpec) Element {
	return SyslogEncoding{
		ComponentID: id,
		RFC:         strings.ToLower(string(o.Syslog.RFC)),
		Facility:    Facility(o.Syslog),
		Severity:    Severity(o.Syslog),
		AppName:     AppName(o.Syslog),
		ProcID:      ProcID(o.Syslog),
		MsgID:       MsgID(o.Syslog),
		PayloadKey:  PayloadKey(o.Syslog),
	}
}

func Facility(s *obs.Syslog) string {
	if s == nil || s.Facility == "" {
		return "user"
	}
	if IsKeyExpr(s.Facility) {
		return fmt.Sprintf("$%s", s.Facility)
	}
	return s.Facility
}

func Severity(s *obs.Syslog) string {
	if s == nil || s.Severity == "" {
		return "informational"
	}
	if IsKeyExpr(s.Severity) {
		return fmt.Sprintf("$%s", s.Severity)
	}
	return s.Severity
}

func AppName(s *obs.Syslog) Element {
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

func MsgID(s *obs.Syslog) Element {
	if s == nil || s.MsgID == "" {
		return Nil
	}
	msgid := "msg_id"
	if IsKeyExpr(s.MsgID) {
		return KV(msgid, fmt.Sprintf(`"$%s"`, s.MsgID))
	}
	return KV(msgid, fmt.Sprintf(`"%s"`, s.MsgID))
}

func ProcID(s *obs.Syslog) Element {
	if s == nil || s.ProcID == "" {
		return Nil
	}
	procid := "proc_id"
	if IsKeyExpr(s.ProcID) {
		return KV(procid, fmt.Sprintf(`"$%s"`, s.ProcID))
	}
	return KV(procid, fmt.Sprintf(`"%s"`, s.ProcID))
}

func PayloadKey(s *obs.Syslog) Element {
	if s == nil || s.PayloadKey == "" {
		return Nil
	}
	key := "payload_key"
	if IsKeyExpr(s.PayloadKey) {
		return KV(key, fmt.Sprintf(`"$%s"`, s.PayloadKey))
	}
	return KV(key, fmt.Sprintf(`"%s"`, s.PayloadKey))
}

// The Syslog output fields can be set to an expression of the form $.abc.xyz
// If an expression is used, its value will be taken from corresponding key in the record
// Example: $.message.procid_key
var keyre = regexp.MustCompile(`^\$(\.[[:word:]]*)+$`)

func IsKeyExpr(str string) bool {
	return keyre.MatchString(str)
}
