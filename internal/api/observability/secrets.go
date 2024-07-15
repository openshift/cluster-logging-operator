package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

// NewSecretReference returns a SecretReference with the given key name and secret
func NewSecretReference(keyName, secretName string) *obs.SecretReference {
	return &obs.SecretReference{
		Key:        keyName,
		SecretName: secretName,
	}
}
