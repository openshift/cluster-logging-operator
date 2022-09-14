package syslog

import (
	"fmt"
	"regexp"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
)

const SyslogHostnameVerify = "syslog_hostname_verify"

type Syslog struct {
	Desc           string
	StoreID        string
	Host           string
	Port           string
	Rfc            string
	Facility       string
	Severity       string
	AppName        Element
	MsgID          Element
	ProcID         Element
	Tag            Element
	Protocol       string
	PayloadKey     string
	SecurityConfig []Element
	BufferConfig   []Element
}

func (s Syslog) Name() string {
	return "syslogTemplate"
}

func (s Syslog) Template() string {
	return `{{define "` + s.Name() + `" -}}
@type remote_syslog
@id {{.StoreID}}
host {{.Host}}
port {{.Port}}
rfc {{.Rfc}}
facility {{.Facility}}
severity {{.Severity}}
{{optional .AppName -}}
{{optional .MsgID -}}
{{optional .ProcID -}}
{{optional .Tag -}}
protocol {{.Protocol}}
packet_size 4096
hostname "#{ENV['NODE_NAME']}"
{{if .SecurityConfig -}}
{{compose .SecurityConfig}}
{{end -}}
{{if (eq .Protocol "tcp") -}}
timeout 60
timeout_exception true
keep_alive true
keep_alive_idle 75
keep_alive_cnt 9
keep_alive_intvl 7200
{{end -}}
{{if .PayloadKey -}}
<format>
  @type single_json_value
  message_key {{.PayloadKey}}
</format>
{{else -}}
<format>
  @type json
</format>
{{end -}}
{{compose .BufferConfig}}
{{end}}
`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	if _, ok := op[generator.UseOldRemoteSyslogPlugin]; ok {
		return ConfOld(bufspec, secret, o, op)
	}
	var addLogSource Element = nil
	var sendJournalLogs Element = nil
	if o.Syslog != nil && o.Syslog.AddLogSource {
		addLogSource = AddLogSource(o, op)
		sendJournalLogs = OutputConf(bufspec, secret, o, source.JournalTags, op)
	}
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				ParseJson(o, op),
				addLogSource,
				sendJournalLogs,
				OutputConf(bufspec, secret, o, "**", op),
			},
		},
	}
}

func OutputConf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, tags string, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	// URL is parsable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = ""
	}
	storeID := helpers.StoreID("", o.Name, "")
	if tags == source.JournalTags {
		storeID = helpers.StoreID("", o.Name, "_journal")
	}
	bufKeys := BufferKeys(o.Syslog, tags)
	return Match{
		MatchTags: tags,
		MatchElement: Syslog{
			StoreID:        storeID,
			Host:           u.Hostname(),
			Port:           port,
			Rfc:            Rfc(o.Syslog),
			Facility:       Facility(o.Syslog),
			Severity:       Severity(o.Syslog),
			AppName:        AppName(o.Syslog, tags),
			MsgID:          MsgID(o.Syslog),
			ProcID:         ProcID(o.Syslog),
			Tag:            Tag(o.Syslog, tags),
			Protocol:       Protocol(o),
			PayloadKey:     PayloadKey(o.Syslog),
			SecurityConfig: SecurityConfig(o, secret),
			BufferConfig:   output.Buffer(bufKeys, bufspec, storeID, &o),
		},
	}
}

func Facility(s *logging.Syslog) string {
	if s == nil || s.Facility == "" {
		return "user"
	}
	if IsKeyExpr(s.Facility) {
		return fmt.Sprintf("${%s}", s.Facility)
	}
	return s.Facility
}

func Severity(s *logging.Syslog) string {
	if s == nil || s.Severity == "" {
		return "debug"
	}
	if IsKeyExpr(s.Severity) {
		return fmt.Sprintf("${%s}", s.Severity)
	}
	return s.Severity
}

func Rfc(s *logging.Syslog) string {
	if s == nil || s.RFC == "" {
		return "rfc5424"
	}
	switch strings.ToLower(s.RFC) {
	case "rfc3164":
		return "rfc3164"
	case "rfc5424":
		return "rfc5424"
	}
	return "Unknown Rfc"
}

func AppName(s *logging.Syslog, matchtags string) Element {
	if s == nil {
		return Nil
	}
	appname := "appname"
	if matchtags == source.JournalTags && Rfc(s) == "rfc5424" {
		return KV(appname, "${$.systemd.u.SYSLOG_IDENTIFIER}")
	}
	if s.AppName == "" {
		return Nil
	}
	if IsKeyExpr(s.AppName) {
		return KV(appname, fmt.Sprintf("${%s}", s.AppName))
	}
	if IsTagExpr(s.AppName) {
		return KV(appname, s.AppName)
	}
	if s.AppName == "tag" {
		return KV(appname, "${tag}")
	}
	return KV(appname, s.AppName)
}

func Tag(s *logging.Syslog, matchtags string) Element {
	if s == nil {
		return Nil
	}
	program := "program"
	if matchtags == source.JournalTags && Rfc(s) == "rfc3164" {
		return KV(program, "${$.systemd.u.SYSLOG_IDENTIFIER}")
	}
	if s.Tag == "" {
		return Nil
	}
	if IsKeyExpr(s.Tag) {
		return KV(program, fmt.Sprintf("${%s}", s.Tag))
	}
	if IsTagExpr(s.Tag) {
		return KV(program, s.Tag)
	}
	if s.Tag == "tag" {
		return KV(program, "${tag}")
	}
	return KV(program, s.Tag)
}

func MsgID(s *logging.Syslog) Element {
	if s == nil || s.MsgID == "" {
		return Nil
	}
	msgid := "msgid"
	if IsKeyExpr(s.MsgID) {
		return KV(msgid, fmt.Sprintf("${%s}", s.MsgID))
	}
	return KV(msgid, s.MsgID)
}

func ProcID(s *logging.Syslog) Element {
	if s == nil || s.ProcID == "" {
		return Nil
	}
	procid := "procid"
	if IsKeyExpr(s.ProcID) {
		return KV(procid, fmt.Sprintf("${%s}", s.ProcID))
	}
	return KV(procid, s.ProcID)
}

func Protocol(o logging.OutputSpec) string {
	u, _ := urlhelper.Parse(o.URL)
	return urlhelper.PlainScheme(u.Scheme)
}

func PayloadKey(s *logging.Syslog) string {
	if s != nil {
		return s.PayloadKey
	}
	return ""
}

func BufferKeys(s *logging.Syslog, matchtags string) []string {
	if s == nil {
		return output.NOKEYS
	}
	keys := []string{}
	tagAdded := false
	if matchtags == source.JournalTags {
		keys = append(keys, "$.systemd.u.SYSLOG_IDENTIFIER")
	}
	if IsKeyExpr(s.Tag) {
		keys = append(keys, s.Tag)
	}
	if IsTagExpr(s.Tag) && !tagAdded {
		keys = append(keys, "tag")
		tagAdded = true
	}
	if s.Tag == "tag" && !tagAdded {
		keys = append(keys, "tag")
		tagAdded = true
	}
	if IsKeyExpr(s.AppName) {
		keys = append(keys, s.AppName)
	}
	if IsTagExpr(s.AppName) && !tagAdded {
		keys = append(keys, "tag")
		tagAdded = true
	}
	if s.AppName == "tag" && !tagAdded {
		keys = append(keys, "tag")
	}
	if IsKeyExpr(s.MsgID) {
		keys = append(keys, s.MsgID)
	}
	if IsKeyExpr(s.ProcID) {
		keys = append(keys, s.ProcID)
	}
	if IsKeyExpr(s.Facility) {
		keys = append(keys, s.Facility)
	}
	if IsKeyExpr(s.Severity) {
		keys = append(keys, s.Severity)
	}
	return keys
}

func ParseJson(o logging.OutputSpec, op Options) Element {
	return Filter{
		MatchTags: "**",
		Element: ConfLiteral{
			TemplateName: "syslogParseJson",
			TemplateStr: `
{{define "syslogParseJson" -}}
@type parse_json_field
json_fields  message
merge_json_log false
replace_json_log true
{{end}}`,
		},
	}
}

func AddLogSource(o logging.OutputSpec, op Options) Element {
	return Filter{
		MatchTags: "**",
		Element: RecordModifier{
			Records: []Record{
				{
					Key:        "kubernetes_info",
					Expression: `${if record.has_key?('kubernetes'); record['kubernetes']; else {}; end}`,
				},
				{
					Key:        "namespace_info",
					Expression: `${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "namespace_name=" + record['kubernetes_info']['namespace_name']; else nil; end}`,
				},
				{
					Key:        "pod_info",
					Expression: `${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "pod_name=" + record['kubernetes_info']['pod_name']; else nil; end}`,
				},
				{
					Key:        "container_info",
					Expression: `${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "container_name=" + record['kubernetes_info']['container_name']; else nil; end}`,
				},
				{
					Key:        "msg_key",
					Expression: `${if record.has_key?('message') && record['message'] != nil; record['message']; else nil; end}`,
				},
				{
					Key:        "msg_info",
					Expression: `${if record['msg_key'] != nil && record['msg_key'].is_a?(Hash); require 'json'; "message="+record['message'].to_json; elsif record['msg_key'] != nil; "message="+record['message']; else nil; end}`,
				},
				{
					Key:        "message",
					Expression: `${if record['msg_key'] != nil && record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; record['namespace_info'] + ", " + record['container_info'] + ", " + record['pod_info'] + ", " + record['msg_info']; else record['message']; end}`,
				},
				{
					Key:        "systemd_info",
					Expression: `${if record.has_key?('systemd') && record['systemd']['t'].has_key?('PID'); record['systemd']['u']['SYSLOG_IDENTIFIER'] += "[" + record['systemd']['t']['PID'] + "]"; else {}; end}`,
				},
			},
			RemoveKeys: []string{
				"kubernetes_info",
				"namespace_info",
				"pod_info",
				"container_info",
				"msg_key",
				"msg_info",
				"systemd_info",
			},
		},
	}
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	if o.Secret != nil {
		conf := []Element{
			TLS(true),
		}
		verify, ok := security.TryKeys(secret, SyslogHostnameVerify)
		if ok {
			if strings.ToLower(string(verify)) == "true" {
				conf = append(conf, HostnameVerify(true))
			} else {
				conf = append(conf, HostnameVerify(false))
			}
		}
		// if secret does not contain "hostname_verify" key, default behavior is expected
		if secret != nil {
			if security.HasTLSCertAndKey(secret) {
				kc := TLSKeyCert{
					CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
					KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
				}
				conf = append(conf, kc)
			}
			if security.HasCABundle(secret) {
				ca := CAFile{
					CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
				}
				conf = append(conf, ca)
			}
		}
		return conf
	} else {
		return nil
	}
}

// The Syslog output fields can be set to an expression of the form $.abc.xyz
// If an expression is used, its value will be taken from corresponding key in the record
// Example: $.message.procid_key
var keyre = regexp.MustCompile(`^\$(\.[[:word:]]*)+$`)

// Example:
// tag
// $.kubernetes.metadata.namespace_name
var tagre = regexp.MustCompile(`\${tag\[-??\d+\]}`)

func IsKeyExpr(str string) bool {
	return keyre.MatchString(str)
}

func IsTagExpr(str string) bool {
	return tagre.MatchString(str)
}
