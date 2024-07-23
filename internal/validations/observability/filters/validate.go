package filters

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
)

func Validate(context internalcontext.ForwarderContext) {
	filterMap := internalobs.FilterMap(context.Forwarder.Spec)
	for _, filter := range filterMap {
		internalobs.SetCondition(&context.Forwarder.Status.Filters,
			internalobs.NewConditionFromPrefix(obs.ConditionTypeValidFilterPrefix, filter.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("filter %q is valid", filter.Name)))
	}

}
