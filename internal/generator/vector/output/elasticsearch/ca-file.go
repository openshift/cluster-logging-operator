package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "elasticsearchCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
tls.ca_file = {{.CAFilePath}}
{{- end}}
`
}
