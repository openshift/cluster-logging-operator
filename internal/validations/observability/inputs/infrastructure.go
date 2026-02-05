package inputs

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
)

func ValidateInfrastructure(spec obs.InputSpec) []metav1.Condition {
	if spec.Type != obs.InputTypeInfrastructure {
		return nil
	}

	if spec.Infrastructure == nil {
		return []metav1.Condition{
			internalobs.NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil infrastructure spec", spec.Name)),
		}
	}
	if len(spec.Infrastructure.Sources) == 0 {
		return []metav1.Condition{
			internalobs.NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, fmt.Sprintf("%s must define at least one valid source", spec.Name)),
		}
	}
	if !set.New(spec.Infrastructure.Sources...).Has(obs.InfrastructureSourceContainer) && spec.Infrastructure.Tuning != nil &&
		spec.Infrastructure.Tuning.Container != nil && spec.Infrastructure.Tuning.Container.MaxMessageSize != nil {
		sources := make([]string, len(spec.Infrastructure.Sources))
		for i, s := range spec.Infrastructure.Sources {
			sources[i] = string(s)
		}
		strSources := strings.Join(sources, ",")

		return []metav1.Condition{
			internalobs.NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, fmt.Sprintf("%s tuning section available only for \"container\" source type, but found %s", spec.Name, strSources)),
		}
	}
	return []metav1.Condition{
		internalobs.NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, spec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("input %q is valid", spec.Name)),
	}
}
