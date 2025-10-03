package observability

import (
	"fmt"
	"regexp"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

const (
	validMaxUnavailableRegex = `^(100%|[1-9][0-9]?%|[1-9][0-9]*)$`
)

var (
	compiledMaxUnavailableRegex = regexp.MustCompile(validMaxUnavailableRegex)
	allowedLogLevels            = sets.NewString("trace", "debug", "info", "warn", "error", "off")
)

func IsPercentOrWholeNumber(val string) bool {
	return compiledMaxUnavailableRegex.MatchString(val)
}

func validateMaxUnavailableAnnotation(context internalcontext.ForwarderContext) {
	if value, ok := context.Forwarder.Annotations[constants.AnnotationMaxUnavailable]; ok {
		if !IsPercentOrWholeNumber(value) {
			condition := internalobs.NewCondition(obs.ConditionTypeMaxUnavailable, obs.ConditionFalse, obs.ReasonMaxUnavailableSupported, "")
			condition.Message = fmt.Sprintf("max-unavailable-rollout value %q must be an absolute number or a valid percentage", value)
			internalobs.SetCondition(&context.Forwarder.Status.Conditions, condition)
			return
		}
	}
	// Condition is only necessary when it is invalid, otherwise we can remove
	internalobs.RemoveConditionByType(&context.Forwarder.Status.Conditions, obs.ConditionTypeMaxUnavailable)
}

func validateLogLevelAnnotation(context internalcontext.ForwarderContext) {
	if level, ok := context.Forwarder.Annotations[constants.AnnotationVectorLogLevel]; ok {
		if !allowedLogLevels.Has(level) {
			condition := internalobs.NewCondition(obs.ConditionTypeLogLevel, obs.ConditionFalse, obs.ReasonLogLevelSupported, "")
			list := strings.Join(allowedLogLevels.List(), ", ")
			condition.Message = fmt.Sprintf("log level %q must be one of [%s]", level, list)
			internalobs.SetCondition(&context.Forwarder.Status.Conditions, condition)
			return
		}
	}
	// Condition is only necessary when it is invalid, otherwise we can remove
	internalobs.RemoveConditionByType(&context.Forwarder.Status.Conditions, obs.ConditionTypeLogLevel)
}
