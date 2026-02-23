package observability

import (
	"fmt"
	"hash/fnv"
	"sort"

	corev1 "k8s.io/api/core/v1"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

// NewSecretReference returns a SecretReference with the given key name and secret
func NewSecretReference(keyName, secretName string) *obs.SecretReference {
	return &obs.SecretReference{
		Key:        keyName,
		SecretName: secretName,
	}
}

// Secrets is a map of secrets
type Secrets map[string]*corev1.Secret

// Hash64a returns an FNV-1a representation of the secrets
func (s Secrets) Hash64a() string {
	names := s.Names()
	buffer := fnv.New64a()
	for _, name := range names {
		secret := s[name]
		buffer.Write([]byte(name))

		var keys []string
		for key := range secret.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := secret.Data[k]
			buffer.Write([]byte(k))
			buffer.Write(v)
		}
	}
	return fmt.Sprintf("%d", buffer.Sum64())
}

func (s Secrets) Names() (names []string) {
	for name := range s {
		names = append(names, name)
	}

	sort.Strings(names)
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
func (s Secrets) Path(key *obs.SecretReference, formatter ...string) string {
	if s.Value(key) != nil {
		return helpers.SecretPath(key.SecretName, key.Key, formatter...)
	}
	return ""
}
