package observability

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

var vectorLogLevelSet = sets.NewString("trace", "debug", "info", "warn", "error", "off")

func validateAnnotations(context internalcontext.ForwarderContext) {
	// No annotations to validate
	clf := context.Forwarder
	if len(clf.Annotations) == 0 {
		return
	}
	// log level annotation
	if level, ok := clf.Annotations[constants.AnnotationVectorLogLevel]; ok {
		condition := internalobs.NewCondition(obs.ConditionTypeLogLevel, obs.ConditionTrue, obs.ReasonLogLevelSupported, "log level is valid")
		if !vectorLogLevelSet.Has(level) {
			condition.Status = obs.ConditionFalse
			condition.Message = fmt.Sprintf("log level %q must be one of trace, debug, info, warn, error, off.", level)
		}
		internalobs.SetCondition(&clf.Status.Conditions, condition)
	}
}
