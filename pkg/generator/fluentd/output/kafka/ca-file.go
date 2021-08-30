package kafka

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "kafkaCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
ssl_ca_cert {{.CAFilePath}}
{{- end}}
`
}
