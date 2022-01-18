package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "elasticsearchCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
ca_file = {{.CAFilePath}}
{{- end}}
`
}
