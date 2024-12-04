package common

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

// ValidateValueReference checks for valid names and keys referenced in secrets and configMaps
func ValidateValueReference(configs []*obsv1.ValueReference, secrets map[string]*corev1.Secret, configMaps map[string]*corev1.ConfigMap) (messages []string) {
	if len(configs) == 0 {
		return messages
	}

	for _, entry := range configs {
		if entry.SecretName != "" {
			if secret, found := secrets[entry.SecretName]; found {
				if value, found := secret.Data[entry.Key]; !found {
					messages = append(messages, fmt.Sprintf("secret[%s.%s] not found", entry.SecretName, entry.Key))
				} else {
					if len(value) == 0 {
						messages = append(messages, fmt.Sprintf("secret[%s.%s] value is empty", entry.SecretName, entry.Key))
					}
				}
			} else {
				messages = append(messages, fmt.Sprintf("secret[%s] not found", entry.SecretName))
			}
		} else if entry.ConfigMapName != "" {
			if cm, found := configMaps[entry.ConfigMapName]; found {
				if value, found := cm.Data[entry.Key]; !found {
					messages = append(messages, fmt.Sprintf("configmap[%s.%s] not found", entry.ConfigMapName, entry.Key))
				} else {
					if strings.TrimSpace(value) == "" {
						messages = append(messages, fmt.Sprintf("configmap[%s.%s] value is empty", entry.ConfigMapName, entry.Key))
					}
				}
			} else {
				messages = append(messages, fmt.Sprintf("configmap[%s] not found", entry.ConfigMapName))
			}
		}
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
