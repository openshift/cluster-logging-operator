package capabilities

import (
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/capabilities/log7571networkpolicy"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
)

type reconcileHandler func(context internalcontext.ForwarderContext, objVisitor func(o runtime.Object)) error

var (
	registry map[string]reconcileHandler = map[string]reconcileHandler{
		log7571networkpolicy.Log7571NetworkPolicyName: log7571networkpolicy.Handle,
	}
)

func ReconcileHandlers(capabilities internalcontext.Capabilities) map[string]reconcileHandler {

	enabledHanlers := map[string]reconcileHandler{}
	for key, cap := range capabilities {
		if handler, found := registry[key]; cap.Enabled && found {
			enabledHanlers[key] = handler
		}
	}
	return enabledHanlers
}
