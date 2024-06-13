package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
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
			bt.Token = helpers.SecretFrom(&obs.SecretKey{
				Secret: &corev1.LocalObjectReference{
					Name: key.Secret.Name,
				},
				Key: key.Secret.Key,
			})

		}
	}
	return bt
}
