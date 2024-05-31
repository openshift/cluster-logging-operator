package inputs

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateReceiver validates receiver input specs
func ValidateReceiver(spec obs.InputSpec) []metav1.Condition {
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

	return nil
}
