package inputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
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
	newCond := func(message string) []metav1.Condition {
		return []metav1.Condition{
			internalobs.NewCondition(obs.ValidationCondition,
				metav1.ConditionTrue,
				obs.ReasonValidationFailure,
				message,
			),
		}
	}

	if spec.Receiver == nil {
		return newCond(fmt.Sprintf("%s has nil receiver spec", spec.Name))
	}
	if spec.Receiver.Type == obs.ReceiverTypeHTTP && spec.Receiver.HTTP == nil {
		return newCond(fmt.Sprintf("%s has nil HTTP receiver spec", spec.Name))
	}
	if spec.Receiver.Type == obs.ReceiverTypeHTTP && spec.Receiver.HTTP.Format == "" {
		return newCond(fmt.Sprintf("%s does not specify a format", spec.Name))
	}
	if spec.Receiver.TLS != nil {
		tlsSpec := obs.TLSSpec(*spec.Receiver.TLS)
		return common.ValidateConfigMapOrSecretKey(spec.Name, internalobs.ConfigMapOrSecretKeys(tlsSpec), secrets, configMaps)
	}

	return nil
}
