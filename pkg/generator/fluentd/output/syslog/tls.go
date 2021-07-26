package syslog

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type TLS security.TLS

func (t TLS) Name() string {
	return "syslogTLSTemplate"
}

func (t TLS) Template() string {
	if t {
		return `{{define "syslogTLSTemplate" -}}
tls true
verify_mode true
{{- end}}`
	}
	return `{{define "syslogTLSTemplate" -}}
tls false
{{end}}`
}
