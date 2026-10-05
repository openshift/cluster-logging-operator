package common

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

// tomlLiteralTerminator is the delimiter go-toml v1 uses for literal multiline strings. Values
// that flow into generated VRL (which is serialized as a literal multiline TOML string) must not
// contain it, or they break out of the string and inject arbitrary TOML into the collector
// configuration. Newlines are likewise rejected: these fields are single-line by nature and a
// newline is what an attacker uses to append injected TOML once the string is terminated.
// This is the admission-time counterpart to the structural guard in internal/utils/toml.
// See LOG-9752 (and LOG-9704 for the originally reported drop-filter sink).
const tomlLiteralTerminator = "'''"

// ValidateTOMLLiteralSafe returns a validation message when value would allow TOML literal
// multiline injection once rendered into the generated collector configuration. It returns an
// empty string when the value is safe.
func ValidateTOMLLiteralSafe(fieldName, value string) string {
	if strings.Contains(value, tomlLiteralTerminator) {
		return fmt.Sprintf("%s must not contain the sequence %q", fieldName, tomlLiteralTerminator)
	}
	if strings.ContainsAny(value, "\n\r") {
		return fmt.Sprintf("%s must not contain newlines or carriage returns", fieldName)
	}
	return ""
}

// ValidateValueReference checks for valid names and keys referenced in secrets and configMaps
func ValidateValueReference(configs []*obsv1.ValueReference, secrets map[string]*corev1.Secret, configMaps map[string]*corev1.ConfigMap) (messages []string) {
	for _, entry := range configs {
		switch {
		case entry.SecretName != "":
			messages = append(messages, validateSecret(entry.SecretName, entry.Key, secrets)...)
		case entry.ConfigMapName != "":
			messages = append(messages, validateConfigMap(entry.ConfigMapName, entry.Key, configMaps)...)
		}
	}
	return messages
}

func validateSecret(secretName, key string, secrets map[string]*corev1.Secret) (messages []string) {
	secret, found := secrets[secretName]
	if !found {
		return []string{fmt.Sprintf("secret[%s] not found", secretName)}
	}
	if value, keyFound := secret.Data[key]; !keyFound {
		messages = append(messages, fmt.Sprintf("secret[%s.%s] not found", secretName, key))
	} else if len(value) == 0 {
		messages = append(messages, fmt.Sprintf("secret[%s.%s] value is empty", secretName, key))
	}
	return messages
}

func validateConfigMap(configMapName, key string, configMaps map[string]*corev1.ConfigMap) (messages []string) {
	cm, found := configMaps[configMapName]
	if !found {
		return []string{fmt.Sprintf("configmap[%s] not found", configMapName)}
	}
	if value, keyFound := cm.Data[key]; !keyFound {
		messages = append(messages, fmt.Sprintf("configmap[%s.%s] not found", configMapName, key))
	} else if strings.TrimSpace(value) == "" {
		messages = append(messages, fmt.Sprintf("configmap[%s.%s] value is empty", configMapName, key))
	}
	return messages
}

// IsEnabledAnnotation checks if an annotation is set to either "true" or "enabled"
func IsEnabledAnnotation(context internalcontext.ForwarderContext, annotation string) bool {
	enabledValues := sets.NewString("true", "enabled")
	if value, ok := context.Forwarder.Annotations[annotation]; ok {
		if enabledValues.Has(value) {
			return true
		}
	}
	return false
}
