package common

import (
	"fmt"
	"strings"

	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	corev1 "k8s.io/api/core/v1"
)

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
func IsEnabledAnnotation(forwarder obsv1.ClusterLogForwarder, annotation string) bool {
	enabledValues := sets.NewString("true", "enabled")
	if value, ok := forwarder.Annotations[annotation]; ok {
		if enabledValues.Has(value) {
			return true
		}
	}
	return false
}
