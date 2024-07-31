package common

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	corev1 "k8s.io/api/core/v1"
)

func MapHTTPAuth(secret *corev1.Secret) *obs.HTTPAuthentication {
	httpAuth := obs.HTTPAuthentication{}
	if security.HasUsernamePassword(secret) {
		httpAuth.Username = &obs.SecretReference{
			Key:        constants.ClientUsername,
			SecretName: secret.Name,
		}
		httpAuth.Password = &obs.SecretReference{
			Key:        constants.ClientPassword,
			SecretName: secret.Name,
		}
	}
	if security.HasBearerTokenFileKey(secret) {
		httpAuth.Token = &obs.BearerToken{
			From: obs.BearerTokenFromSecret,
			Secret: &obs.BearerTokenSecretKey{
				Name: secret.Name,
				Key:  constants.BearerTokenFileKey,
			},
		}
	}
	return &httpAuth
}
