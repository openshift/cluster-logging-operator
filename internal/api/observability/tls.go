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
	if t.CA != nil && t.CA.Secret != nil {
		secrets.Insert(t.CA.Secret.Name)
	}
	if t.Certificate != nil && t.Certificate.Secret != nil {
		secrets.Insert(t.Certificate.Secret.Name)
	}
	if t.KeyPassphrase != nil && t.KeyPassphrase.Secret != nil {
		secrets.Insert(t.KeyPassphrase.Secret.Name)
	}
	return secrets.UnsortedList()
}

// ConfigmapsForTLS returns the unique set of configmap names for a TLS spec
func ConfigmapsForTLS(t obsv1.TLSSpec) []string {
	configmaps := set.New[string]()
	if t.CA != nil && t.CA.From == obsv1.ConfigReferenceFromConfigMap {
		configmaps.Insert(t.CA.ConfigMap.Name)
	}
	if t.Certificate != nil && t.Certificate.From == obsv1.ConfigReferenceFromConfigMap {
		configmaps.Insert(t.Certificate.ConfigMap.Name)
	}
	return configmaps.UnsortedList()
}

// ConfigReferences returns a list ConfigReference, converting SecretConfigReferences as necessary
func ConfigReferences(t obsv1.TLSSpec) []*obsv1.ConfigReference {
	results := []*obsv1.ConfigReference{}
	if t.CA != nil {
		results = append(results, t.CA)
	}
	if t.Certificate != nil {
		results = append(results, t.Certificate)
	}
	if t.Key != nil {
		results = append(results, &obsv1.ConfigReference{
			Key:    t.Key.Key,
			Secret: t.Key.Secret,
		})
	}
	if t.KeyPassphrase != nil {
		results = append(results, &obsv1.ConfigReference{
			Key:    t.KeyPassphrase.Key,
			Secret: t.KeyPassphrase.Secret,
		})
	}

	return results
}
