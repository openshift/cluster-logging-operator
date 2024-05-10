package helpers

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	corev1 "k8s.io/api/core/v1"
)

func GetOutputSecret(o logging.OutputSpec, secrets map[string]*corev1.Secret) *corev1.Secret {
	if s, ok := secrets[o.Name]; ok {
		log.V(9).Info("Using secret configured in output: " + o.Name)
		return s
	}
	if o.Type == logging.OutputTypeLoki && lokistack.DefaultLokiOutputNames.Has(o.Name) {
		log.V(9).Info("Default lokiStack, using collector token", "output", o.Name, "secret", constants.LogCollectorToken)
		return secrets[constants.LogCollectorToken]
	}
	log.V(9).Info("No Secret found for output", "output", o.Name)
	return nil
}

// Secrets is a map of secrets
type Secrets map[string]*corev1.Secret

// Value returns the content of the given secret with key if it exists or nil
func (s Secrets) Value(key *v1.SecretKey) []byte {
	if key != nil && key.Secret != nil {
		if secret, exists := s[key.Secret.Name]; exists {
			if value, exists := secret.Data[key.Key]; exists {
				return value
			}
		}
	}
	return nil
}

// AsString returns the value of the given secret with key if it exists or empty
func (s Secrets) AsString(key *v1.SecretKey) string {
	if v := s.Value(key); v != nil {
		return string(v)
	}
	return ""
}

// AsString returns the value of the BeearerToken if it exists or empty
func (s Secrets) AsStringFromBearerToken(key *v1.BearerToken) string {
	if key.Secret != nil {
		return s.AsString(&v1.SecretKey{
			Secret: key.Secret,
			Key:    key.Key,
		})
	}
	// TODO: find value when SA is defined
	return ""
}
