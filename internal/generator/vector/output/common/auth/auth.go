package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

// HTTPAuth provides auth configuration for http authentication where username/password or bearer token
// are viable options.  Bearer token takes precedence if both are provided
func HTTPAuth(id string, spec *obs.HTTPAuthentication, secrets vectorhelpers.Secrets) Element {
	if spec != nil {
		if spec.Token != nil {
			return NewBearerToken(id, spec, secrets)
		}
		return NewBasic(id, spec, secrets)
	}
	return Nil
}
