package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type CAFile security.CAFile

func (ca CAFile) Name() string {
	return "fluentdforwardCAFileTemplate"
}

func (ca CAFile) Template() string {
	return `{{define "` + ca.Name() + `" -}}
tls_cert_path {{.CAFilePath}}
{{- end}}
`
}
