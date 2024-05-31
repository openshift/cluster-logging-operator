package inputs

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

var (
	globRE = regexp.MustCompile(`^[a-zA-Z0-9\*\-]*$`)

	globErrorFmt = `input %q: invalid glob for %s. Must match '` + globRE.String() + `"`
)

func ValidateApplication(spec obs.InputSpec) []metav1.Condition {
	if spec.Type != obs.InputTypeApplication {
		return nil
	}
	newCond := func(reason, message string, args ...any) metav1.Condition {
		if len(args) > 1 {
			message = fmt.Sprintf(message, args...)
		}
		return internalobs.NewCondition(obs.ValidationCondition,
			metav1.ConditionTrue,
			reason,
			message,
		)
	}

	if spec.Application == nil {
		return []metav1.Condition{newCond(obs.ReasonMissingSpec, fmt.Sprintf("%s has nil application spec", spec.Name))}
	}
	conditions := []metav1.Condition{}
	switch {
	case spec.Application.Excludes != nil:
		for _, ex := range spec.Application.Excludes {
			if !globRE.MatchString(ex.Namespace) {
				conditions = append(conditions, newCond(obs.ReasonInvalidGlob, globErrorFmt, spec.Name, "namespace excludes"))
			}
			if !globRE.MatchString(ex.Container) {
				conditions = append(conditions, newCond(obs.ReasonInvalidGlob, globErrorFmt, spec.Name, "container excludes"))
			}
		}
	case spec.Application.Includes != nil:
		for _, in := range spec.Application.Includes {
			if !globRE.MatchString(in.Namespace) {
				conditions = append(conditions, newCond(obs.ReasonInvalidGlob, globErrorFmt, spec.Name, "namespace includes"))
			}
			if !globRE.MatchString(in.Container) {
				conditions = append(conditions, newCond(obs.ReasonInvalidGlob, globErrorFmt, spec.Name, "container includes"))
			}
		}
	}
	return conditions
}
