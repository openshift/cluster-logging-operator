package observability

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/utils/set"
)

// SecretsForTLS returns the unique set of secret names for a TLS spec
func SecretsForTLS(t obsv1.TLSSpec) []string {
	secrets := set.New[string]()
	if t.Key != nil {
		secrets.Insert(t.Key.SecretName)
	}
	if t.CA != nil && t.CA.SecretName != "" {
		secrets.Insert(t.CA.SecretName)
	}
	if t.Certificate != nil && t.Certificate.SecretName != "" {
		secrets.Insert(t.Certificate.SecretName)
	}
	if t.KeyPassphrase != nil && t.KeyPassphrase.SecretName != "" {
		secrets.Insert(t.KeyPassphrase.SecretName)
	}
	return secrets.UnsortedList()
}

// ConfigmapsForTLS returns the unique set of configmap names for a TLS spec
func ConfigmapsForTLS(t obsv1.TLSSpec) []string {
	configmaps := set.New[string]()
	if t.CA != nil && t.CA.SecretName == "" && t.CA.ConfigMapName != "" {
		configmaps.Insert(t.CA.ConfigMapName)
	}
	if t.Certificate != nil && t.Certificate.SecretName == "" && t.Certificate.ConfigMapName != "" {
		configmaps.Insert(t.Certificate.ConfigMapName)
	}
	return configmaps.UnsortedList()
}

// ValueReferences returns a slice of ValueReferences, converting SecretReferences as necessary
func ValueReferences(t obsv1.TLSSpec) []*obsv1.ValueReference {
	results := []*obsv1.ValueReference{}
	if t.CA != nil {
		results = append(results, t.CA)
	}
	if t.Certificate != nil {
		results = append(results, t.Certificate)
	}
	if t.Key != nil {
		results = append(results, &obsv1.ValueReference{
			Key:        t.Key.Key,
			SecretName: t.Key.SecretName,
		})
	}
	if t.KeyPassphrase != nil {
		results = append(results, &obsv1.ValueReference{
			Key:        t.KeyPassphrase.Key,
			SecretName: t.KeyPassphrase.SecretName,
		})
	}

	return results
}
