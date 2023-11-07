package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

type UserNamePass common.UserNamePass

func (up UserNamePass) Name() string {
	return "elasticsearchUsernamePasswordTemplate"
}

func (up UserNamePass) Template() string {
	return `{{define "` + up.Name() + `" -}}
strategy = "basic"
user = "{{.Username}}"
password = "{{.Password}}"
{{- end}}`
}
