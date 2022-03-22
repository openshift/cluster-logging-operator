package loki

import (
	security2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type UserNamePass security2.UserNamePass

func (up UserNamePass) Name() string {
	return "lokiUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
username "#{File.read({{ .UsernamePath }}) rescue nil}"
password "#{File.read({{ .PasswordPath }}) rescue nil}"
{{- end}}
`
}
