package inputs

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateAudit(spec obs.InputSpec) []metav1.Condition {
	if spec.Type != obs.InputTypeAudit {
		return nil
	}

	if spec.Audit == nil {
		return []metav1.Condition{
			internalobs.NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil audit spec", spec.Name)),
		}
	}
	return []metav1.Condition{
		internalobs.NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("input %q is valid", spec.Name)),
	}
}
