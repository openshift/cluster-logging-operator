package auth

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/api/observability"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

// HTTPAuth provides auth configuration for http authentication where username/password or bearer token
// are viable options.  Bearer token takes precedence if both are provided
func HTTPAuth(id string, spec *obs.HTTPAuthentication, secrets vectorhelpers.Secrets, op Options) Element {
	if spec != nil {
		if spec.Token != nil {
			return NewBearerToken(id, spec, secrets, op)
		}
		return NewBasic(id, spec, secrets)
	}
	return Nil
}
