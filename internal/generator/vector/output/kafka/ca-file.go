package kafka

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "kafkaCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
ca_file = {{.CAFilePath}}
{{end}}
`
}
