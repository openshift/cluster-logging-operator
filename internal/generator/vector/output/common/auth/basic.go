package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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

func NewBasic(id string, spec *obs.HTTPAuthentication, secrets helpers.Secrets) Basic {
	b := Basic{}
	if spec != nil {
		b.ID = id
		if spec.Username != nil {
			b.Username = secrets.AsString(spec.Username)
		}
		if spec.Password != nil {
			b.Password = secrets.AsString(spec.Password)
		}
	}
	return b
}
