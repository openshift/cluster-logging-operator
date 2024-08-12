package helpers

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

// Secrets is a map of secrets
type Secrets map[string]*corev1.Secret

func (s Secrets) Names() (names []string) {

	for name := range s {
		names = append(names, name)
	}
	return names
}

// Value returns the content of the given secret with key if it exists or nil
func (s Secrets) Value(key *obs.SecretReference) []byte {
	if key != nil && key.SecretName != "" {
		if secret, exists := s[key.SecretName]; exists {
			if value, exists := secret.Data[key.Key]; exists {
				return value
			}
		}
	}
	return nil
}

// AsString returns the value of the given secret with key if it exists or empty
func (s Secrets) AsString(key *obs.SecretReference) string {
	if v := s.Value(key); v != nil {
		return string(v)
	}
	return ""
}

// AsStringFromBearerToken returns the value of the BearerToken if it exists or empty
func (s Secrets) AsStringFromBearerToken(key *obs.BearerToken) string {
	if key.From == obs.BearerTokenFromSecret && key.Secret != nil {
		return s.AsString(&obs.SecretReference{
			Key:        key.Secret.Key,
			SecretName: key.Secret.Name,
		})

	}
	return ""
}

// Path returns the path to the given secret key if it exists or empty
func (s Secrets) Path(key *obs.SecretReference) string {
	if s.Value(key) != nil {
		return SecretPath(key.SecretName, key.Key)
	}
	return ""
}
