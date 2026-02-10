package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// HTTPAuth provides auth configuration for http authentication where username/password or bearer token
// are viable options.  Bearer token takes precedence if both are provided
func HTTPAuth(id string, spec *obs.HTTPAuthentication, secrets internalobs.Secrets, op utils.Options) framework.Element {
	if spec != nil {
		if spec.Token != nil {
			return NewBearerToken(id, spec, secrets, op)
		}
		return NewBasic(id, spec, secrets)
	}
	return framework.Nil
}

func NewHttpAuth(spec *obs.HTTPAuthentication, op utils.Options) (auth *sinks.HttpAuth) {
	if spec == nil {
		return nil
	}
	auth = &sinks.HttpAuth{}
	if spec.Token != nil {
		auth.Strategy = sinks.HttpAuthStrategyBearer
		key := spec.Token
		switch key.From {
		case obs.BearerTokenFromSecret:
			if key.Secret != nil {
				auth.Token = helpers.SecretFrom(&obs.SecretReference{
					SecretName: key.Secret.Name,
					Key:        key.Secret.Key,
				})
			}
		case obs.BearerTokenFromServiceAccount:
			if name, found := utils.GetOption[string](op, framework.OptionServiceAccountTokenSecretName, ""); found {
				auth.Token = helpers.SecretFrom(&obs.SecretReference{
					Key:        constants.TokenKey,
					SecretName: name,
				})
			}
		}
	} else {
		auth.Strategy = sinks.HttpAuthStrategyBasic
		auth.User = helpers.SecretFrom(spec.Username)
		auth.Password = helpers.SecretFrom(spec.Password)
	}

	return auth
}
