package syslog

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/normalize"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	corev1 "k8s.io/api/core/v1"
)

type SyslogOld struct {
	StoreID    string
	PluginType string
	RemoteHost string
	Port       string
	Facility   string
	Severity   string
}

func (so SyslogOld) Name() string {
	return "syslogoldTemplate"
}

func (so SyslogOld) Template() string {
	return `{{define "` + so.Name() + `" -}}
@type {{.PluginType}}
@id {{.StoreID}}
remote_syslog {{.RemoteHost}}
port {{.Port}}
hostname "#{ENV['NODE_NAME']}"
facility {{.Facility}}
severity {{.Severity}}
{{end}}
`
}

func ConfOld(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	// URL is parasable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = ""
	}
	storeID := helpers.StoreID("", o.Name, "")
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				normalize.DedotLabels(),
				ParseJson(o, op),
				Match{
					MatchTags: "**",
					MatchElement: SyslogOld{
						StoreID:    storeID,
						PluginType: SyslogPlugin(o),
						RemoteHost: u.Hostname(),
						Port:       port,
						Facility:   "user",
						Severity:   "debug",
					},
				},
			},
		},
	}
}

func SyslogPlugin(o logging.OutputSpec) string {
	if protocol := Protocol(o); protocol == "udp" {
		return "syslog"
	}
	return "syslog_buffered"
}
