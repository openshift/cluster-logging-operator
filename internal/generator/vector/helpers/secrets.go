package helpers

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetOutputSecret(o logging.OutputSpec, secrets map[string]*corev1.Secret) *corev1.Secret {
	if s, ok := secrets[o.Name]; ok {
		log.V(9).Info("Using secret configured in output: " + o.Name)
		return s
	}
	log.V(9).Info("No Secret found for output", "output", o.Name)
	return nil
}

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
