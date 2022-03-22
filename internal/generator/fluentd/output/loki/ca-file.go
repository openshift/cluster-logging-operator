package loki

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "lokiCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
ca_cert {{.CAFilePath}}
{{- end}}
`
}
