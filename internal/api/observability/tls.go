package observability

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/utils/set"
)

// SecretsForTLS returns the unique set of secret names for a TLS spec
func SecretsForTLS(t obsv1.TLSSpec) []string {
	secrets := set.New[string]()
	if t.Key != nil {
		secrets.Insert(t.Key.Secret.Name)
	}
	if t.CA.Secret != nil {
		secrets.Insert(t.CA.Secret.Name)
	}
	if t.Certificate.Secret != nil {
		secrets.Insert(t.Certificate.Secret.Name)
	}
	if t.KeyPassphrase.Secret != nil {
		secrets.Insert(t.KeyPassphrase.Secret.Name)
	}
	return secrets.UnsortedList()
}

// ConfigmapsForTLS returns the unique set of configmap names for a TLS spec
func ConfigmapsForTLS(t obsv1.TLSSpec) []string {
	configmaps := set.New[string]()
	if t.CA.Secret == nil && t.CA.ConfigMap != nil {
		configmaps.Insert(t.CA.ConfigMap.Name)
	}
	if t.Certificate.Secret == nil && t.Certificate.ConfigMap != nil {
		configmaps.Insert(t.Certificate.ConfigMap.Name)
	}
	return configmaps.UnsortedList()
}
