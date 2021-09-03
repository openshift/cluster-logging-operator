package kafka

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type UserNamePass security.UserNamePass

func (up UserNamePass) Name() string {
	return "kafkaUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
sasl_plain_username "#{File.exists?({{.UsernamePath}}) ? open({{.UsernamePath}},'r') do |f|f.read end : ''}"
sasl_plain_password "#{File.exists?({{.PasswordPath}}) ? open({{.PasswordPath}},'r') do |f|f.read end : ''}"
{{- end}}
`
}
