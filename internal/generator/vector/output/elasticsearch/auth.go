package elasticsearch

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

type BearerToken struct {
	ID    string
	Token string
}

func (bt BearerToken) Name() string {
	return "esBearerTokenTemplate"
}

func (bt BearerToken) Template() string {
	return `{{define "` + bt.Name() + `" -}}
[sinks.{{.ID}}.request.headers]
Authorization = "Bearer {{.Token}}"
{{end}}
`
}

func NewBearerToken(id string, spec *obs.HTTPAuthentication, secrets observability.Secrets, op framework.Options) BearerToken {
	bt := BearerToken{}
	if spec != nil {
		key := spec.Token
		bt.ID = id
		switch key.From {
		case obs.BearerTokenFromSecret:
			if key.Secret != nil {
				bt.Token = helpers.SecretFrom(&obs.SecretReference{
					SecretName: key.Secret.Name,
					Key:        key.Secret.Key,
				})
			}
		case obs.BearerTokenFromServiceAccount:
			if name, found := utils.GetOption(op, framework.OptionServiceAccountTokenSecretName, ""); found {
				bt.Token = helpers.SecretFrom(&obs.SecretReference{
					Key:        constants.TokenKey,
					SecretName: name,
				})
			}
		}
	}
	return bt
}
