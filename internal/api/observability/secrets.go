package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

// NewSecretKey returns a SecretConfigReference with the given key name and secret
func NewSecretKey(keyName, secretName string) *obs.SecretConfigReference {
	return &obs.SecretConfigReference{
		Key: keyName,
		Secret: &corev1.LocalObjectReference{
			Name: secretName,
		},
	}
}
