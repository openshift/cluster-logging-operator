package loki

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type BearerTokenFile security.BearerTokenFile

func (bt BearerTokenFile) Name() string {
	return "lokiBearerTokenFileTemplate"
}

func (bt BearerTokenFile) Template() string {
	return `{{define "` + bt.Name() + `" -}}
bearer_token_file {{ .BearerTokenFilePath }}
{{- end}}
`
}
