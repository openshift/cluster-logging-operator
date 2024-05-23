package helpers

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
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

// Value returns the content of the given secret with key if it exists or nil
func (s Secrets) Value(key *obs.SecretKey) []byte {
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
func (s Secrets) AsString(key *obs.SecretKey) string {
	if v := s.Value(key); v != nil {
		return string(v)
	}
	return ""
}

// AsString returns the value of the BeearerToken if it exists or empty
func (s Secrets) AsStringFromBearerToken(key *obs.BearerToken) string {
	if key.Secret != nil {
		return s.AsString(&obs.SecretKey{
			Secret: key.Secret.Secret,
			Key:    key.Secret.Key,
		})

	}
	// We reconcile SA token secret, so name and key will be known
	return s.AsString(&obs.SecretKey{
		Secret: &corev1.LocalObjectReference{
			Name: key.ServiceAccount.Name + "-token",
		},
		Key: constants.TokenKey,
	})
}

// Path returns the path to the given secret key if it exists or empty
func (s Secrets) Path(key *obs.SecretKey) string {
	if s.Value(key) != nil {
		return common.SecretPath(key.Secret.Name, key.Key)
	}
	return ""
}
