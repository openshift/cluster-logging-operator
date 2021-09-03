package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type UserNamePass security.UserNamePass

func (up UserNamePass) Name() string {
	return "forwardUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
username "#{File.exists?('{{.UsernamePath}}') ? open('{{.UsernamePath}}','r') do |f|f.read end : ''}" 
password "#{File.exists?('{{.PasswordPath}}') ? open('{{.PasswordPath}}','r') do |f|f.read end : ''}"
{{- end}}
`
}
