package observability

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"strings"
)

func validateAnnotations(context internalcontext.ForwarderContext) {
	allowedLogLevel := sets.NewString("trace", "debug", "info", "warn", "error", "off")

	if level, ok := context.Forwarder.Annotations[constants.AnnotationVectorLogLevel]; ok {
		if !allowedLogLevel.Has(level) {
			condition := internalobs.NewCondition(obs.ConditionTypeLogLevel, obs.ConditionFalse, obs.ReasonLogLevelSupported, "")
			list := strings.Join(allowedLogLevel.List(), ", ")
			condition.Message = fmt.Sprintf("log level %q must be one of [%s]", level, list)
			internalobs.SetCondition(&context.Forwarder.Status.Conditions, condition)
			return
		}
	}
	// Condition is only necessary when it is invalid, otherwise we can remove
	internalobs.RemoveConditionByType(&context.Forwarder.Status.Conditions, obs.ConditionTypeLogLevel)
}
