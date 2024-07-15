package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type BearerToken struct {
	ID    string
	Token string
}

func (bt BearerToken) Name() string {
	return "httpBearerTokenTemplate"
}

func (bt BearerToken) Template() string {
	return `{{define "` + bt.Name() + `" -}}
[sinks.{{.ID}}.auth]
strategy = "bearer"
token = "{{.Token}}"
{{end}}
`
}

func NewBearerToken(id string, spec *obs.HTTPAuthentication, secrets helpers.Secrets) BearerToken {
	bt := BearerToken{}
	if spec != nil {
		key := spec.Token
		bt.ID = id
		if key.From == obs.BearerTokenFromSecret && key.Secret != nil {
			bt.Token = helpers.SecretFrom(&obs.SecretReference{
				Key:        key.Secret.Key,
				SecretName: key.Secret.Name,
			})

		}
	}
	return bt
}
