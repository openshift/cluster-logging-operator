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
username = "" # inject code to read from filename
password = "" # inject code to read from filename
{{- end}}
`
}
