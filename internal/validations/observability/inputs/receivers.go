package inputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)

// ValidateReceiver validates receiver input specs
func ValidateReceiver(spec obs.InputSpec, secrets map[string]*corev1.Secret, configMaps map[string]*corev1.ConfigMap) []metav1.Condition {
	if secrets == nil || configMaps == nil {
		log.WithName("ValidateReceier").V(0).Info("runtime error: expected maps of secrets and configmaps")
		os.Exit(1)
	}
	if spec.Type != obs.InputTypeReceiver {
		return nil
	}

	if spec.Receiver == nil {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil receiver spec", spec.Name)),
		}
	}
	if spec.Receiver.Type == obs.ReceiverTypeHTTP && spec.Receiver.HTTP == nil {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil HTTP receiver spec", spec.Name)),
		}
	}
	if spec.Receiver.Type == obs.ReceiverTypeHTTP && spec.Receiver.HTTP.Format == "" {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, fmt.Sprintf("%s does not specify a format", spec.Name)),
		}
	}
	if spec.Receiver.TLS != nil {
		tlsSpec := obs.TLSSpec(*spec.Receiver.TLS)
		if messages := common.ValidateConfigMapOrSecretKey(ConfigMapOrSecretKeys(tlsSpec), secrets, configMaps); len(messages) > 0 {
			return []metav1.Condition{
				NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, strings.Join(messages, ",")),
			}
		}
	}

	return []metav1.Condition{
		NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("input %q is valid", spec.Name)),
	}
}
