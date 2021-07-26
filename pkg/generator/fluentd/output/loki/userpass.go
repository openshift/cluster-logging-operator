package loki

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type UserNamePass security.UserNamePass

func (up UserNamePass) Name() string {
	return "elasticsearchUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
username "#{File.read({{ .UsernamePath }}) rescue nil}"
password "#{File.read({{ .PasswordPath }}) rescue nil}"
{{- end}}
`
}
