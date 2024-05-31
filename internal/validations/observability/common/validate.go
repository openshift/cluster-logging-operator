package common

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateConfigMapOrSecretKey(name string, configs []*obsv1.ConfigMapOrSecretKey, secrets map[string]*corev1.Secret, configMaps map[string]*corev1.ConfigMap) (conditions []metav1.Condition) {
	if len(configs) == 0 {
		return nil
	}

	addCondition := func(reason, message string) {
		conditions = append(conditions, internalobs.NewCondition(obsv1.ValidationCondition,
			metav1.ConditionTrue,
			reason,
			message,
		))
	}

	for _, entry := range configs {
		if entry.Secret != nil {
			if secret, found := secrets[entry.Secret.Name]; found {
				if _, found := secret.Data[entry.Key]; !found {
					addCondition(obsv1.ReasonSecretKeyNotFound, fmt.Sprintf("key %q not found in secret %q for %q", entry.Key, entry.Secret.Name, name))
				}
			} else {
				addCondition(obsv1.ReasonSecretNotFound, fmt.Sprintf("secret %q not found for %q", entry.Secret.Name, name))
			}
		} else if entry.ConfigMap != nil {
			if cm, found := configMaps[entry.ConfigMap.Name]; found {
				if _, found := cm.Data[entry.Key]; !found {
					addCondition(obsv1.ReasonConfigMapKeyNotFound, fmt.Sprintf("key %q not found in configmap %q for %q", entry.Key, entry.ConfigMap.Name, name))
				}
			} else {
				addCondition(obsv1.ReasonConfigMapNotFound, fmt.Sprintf("configmap %q not found for %q", entry.ConfigMap.Name, name))
			}
		}
	}

	return conditions
}
