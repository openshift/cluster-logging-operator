package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

// NewSecretKey returns a SecretKey with the given key name and secret
func NewSecretKey(keyName, secretName string) *obs.SecretKey {
	return &obs.SecretKey{
		Key: keyName,
		Secret: &corev1.LocalObjectReference{
			Name: secretName,
		},
	}
}
