package inputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/golang-collections/collections/set"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
	obsmigrate "github.com/openshift/cluster-logging-operator/internal/migrations/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)

// ValidateReceiver validates receiver input specs
func ValidateReceiver(spec obs.InputSpec, secrets map[string]*corev1.Secret, configMaps map[string]*corev1.ConfigMap, context utils.Options) []metav1.Condition {
	if secrets == nil || configMaps == nil {
		log.WithName("ValidateReceier").V(0).Info("runtime error: expected maps of secrets and configmaps")
		os.Exit(1)
	}
	if spec.Type != obs.InputTypeReceiver {
		return nil
	}

	if spec.Receiver == nil {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil receiver spec", spec.Name)),
		}
	}
	if spec.Receiver.Type == obs.ReceiverTypeHTTP && spec.Receiver.HTTP == nil {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil HTTP receiver spec", spec.Name)),
		}
	}
	if spec.Receiver.Type == obs.ReceiverTypeHTTP && spec.Receiver.HTTP.Format == "" {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, fmt.Sprintf("%s does not specify a format", spec.Name)),
		}
	}
	if spec.Receiver.TLS != nil {
		tlsSpec := obs.TLSSpec(*spec.Receiver.TLS)
		keys := ConfigMapOrSecretKeys(tlsSpec)
		skipKeys := extractSecretKeysAsSet(context)
		keys = removeGeneratedSecrets(keys, skipKeys)
		if messages := common.ValidateConfigMapOrSecretKey(keys, secrets, configMaps); len(messages) > 0 {
			return []metav1.Condition{
				NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, strings.Join(messages, ",")),
			}
		}
	}

	return []metav1.Condition{
		NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("input %q is valid", spec.Name)),
	}
}

func removeGeneratedSecrets(keys []*obs.ConfigMapOrSecretKey, skipKeys *set.Set) (result []*obs.ConfigMapOrSecretKey) {
	for _, secretKey := range keys {
		if secretKey.Secret != nil {
			key := fmt.Sprintf("%v_%v", secretKey.Secret.Name, secretKey.Key)
			if !skipKeys.Has(key) {
				result = append(result, secretKey)
			}
		} else { //configmap
			result = append(result, secretKey)
		}
	}
	return result
}

func extractSecretKeysAsSet(context utils.Options) *set.Set {
	secretKeys := set.New()
	if generatedSecrets, found := utils.GetOption[[]*corev1.Secret](context, obsmigrate.GeneratedSecrets, []*corev1.Secret{}); found {
		for _, secret := range generatedSecrets {
			for key := range secret.Data {
				secretKeys.Insert(fmt.Sprintf("%s_%s", secret.Name, key))
			}
		}
	}
	return secretKeys
}
