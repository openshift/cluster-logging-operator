package common

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

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
