package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type Basic struct {
	ID       string
	Username string
	Password string
}

func isEmpty(b Basic) bool {
	return b == Basic{}
}

func (b Basic) Name() string {
	return "userpasswdTemplate"
}

func (b Basic) Template() string {
	if isEmpty(b) {
		return `{{define "` + b.Name() + `" -}}{{end}}`
	}
	return `{{define "` + b.Name() + `" -}}
[sinks.{{.ID}}.auth]
strategy = "basic"
user = "{{.Username}}"
password = "{{.Password}}"
{{- end}}`
}

func NewBasic(id string, spec *obs.HTTPAuthentication, secrets observability.Secrets) Basic {
	b := Basic{}
	if spec != nil {
		b.ID = id
		b.Username = helpers.SecretFrom(spec.Username)
		b.Password = helpers.SecretFrom(spec.Password)
	}
	return b
}
