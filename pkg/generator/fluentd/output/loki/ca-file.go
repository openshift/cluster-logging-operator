package loki

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "elasticsearchCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
ca_cert {{.CAFilePath}}
{{- end}}
`
}
