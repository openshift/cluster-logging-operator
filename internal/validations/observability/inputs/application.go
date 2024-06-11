package inputs

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
)

var (
	globRE = regexp.MustCompile(`^[a-zA-Z0-9\*\-]*$`)

	globErrorFmt = `input %q: invalid glob for %s. Must match '` + globRE.String() + `"`
)

func ValidateApplication(spec obs.InputSpec) (conditions []metav1.Condition) {
	if spec.Type != obs.InputTypeApplication {
		return nil
	}

	if spec.Application == nil {
		return []metav1.Condition{
			NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, false, obs.ReasonMissingSpec, fmt.Sprintf("%s has nil application spec", spec.Name)),
		}
	}
	var messages []string
	if spec.Application.Excludes != nil {
		for i, ex := range spec.Application.Excludes {
			if !globRE.MatchString(ex.Namespace) {
				messages = append(messages, fmt.Sprintf("excludes[%d].namespace", i))
			}
			if !globRE.MatchString(ex.Container) {
				messages = append(messages, fmt.Sprintf("excludes[%d].container", i))
			}
		}
	}
	if spec.Application.Includes != nil {
		for i, in := range spec.Application.Includes {
			if !globRE.MatchString(in.Namespace) {
				messages = append(messages, fmt.Sprintf("includes[%d].namespace", i))
			}
			if !globRE.MatchString(in.Container) {
				messages = append(messages, fmt.Sprintf("includes[%d].container", i))
			}
		}
	}
	if len(messages) > 0 {
		msg := fmt.Sprintf("globs must match %q for: %s", globRE, strings.Join(messages, ","))
		conditions = append(conditions, NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, false, obs.ReasonValidationFailure, msg))
	} else {
		conditions = append(conditions, NewConditionFromPrefix(obs.ConditionValidInputPrefix, spec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("input %q is valid", spec.Name)))
	}

	return conditions
}
