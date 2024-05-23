package utils

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// GatherSources collects the set of unique source types and namespaces
func GatherSources(forwarder *logging.ClusterLogForwarderSpec) sets.String {
	types := sets.NewString()
	specs := forwarder.InputMap()
	for inputName := range logging.NewRoutes(forwarder.Pipelines).ByInput {
		if logging.ReservedInputNames.Has(inputName) {
			types.Insert(inputName) // Use name as type.
		} else if spec, ok := specs[inputName]; ok {
			if spec.Application != nil {
				types.Insert(logging.InputNameApplication)
			}
			if spec.Infrastructure != nil {
				types.Insert(logging.InputNameInfrastructure)
			}
			if spec.Audit != nil {
				types.Insert(logging.InputNameAudit)
			}
			if spec.Receiver != nil {
				types.Insert(logging.InputNameReceiver)
			}
		}
	}
	return *types
}
