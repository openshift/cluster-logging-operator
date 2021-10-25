package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

type UserNamePass security.UserNamePass

func (up UserNamePass) Name() string {
	return "elasticsearchUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
auth.strategy = "basic"
auth.user = "#{File.exists?({{.UsernamePath}}) ? open({{.UsernamePath}},'r') do |f|f.read end : ''}"
auth.password = "#{File.exists?({{.PasswordPath}}) ? open({{.PasswordPath}},'r') do |f|f.read end : ''}"
{{- end}}
`
}
