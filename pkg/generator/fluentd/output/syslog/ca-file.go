package syslog

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "syslogCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
ca_file {{.CAFilePath}}
{{- end}}
`
}
