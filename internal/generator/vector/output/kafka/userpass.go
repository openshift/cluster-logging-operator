package kafka

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

type UserNamePass common.UserNamePass

func (up UserNamePass) Name() string {
	return "kafkaUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
user = "{{.Username}}"
password = "{{.Password}}"
{{- end}}
`
}
