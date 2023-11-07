package generator

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

const (
	IncludeLegacyForwardConfig = "includeLegacyForwardConfig"
	UseOldRemoteSyslogPlugin   = "useOldRemoteSyslogPlugin"
	ClusterTLSProfileSpec      = "tlsProfileSpec"

	MinTLSVersion = "minTLSVersion"
	Ciphers       = "ciphers"
)

// GatherSources collects the set of unique source types and namespaces
func GatherSources(forwarder *logging.ClusterLogForwarderSpec, op Options) sets.String {
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

func InputsToPipelines(fwdspec *logging.ClusterLogForwarderSpec) logging.RouteMap {
	result := logging.RouteMap{}
	inputs := fwdspec.InputMap()
	for _, pipeline := range fwdspec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if input, ok := inputs[inRef]; ok {
				// User defined input spec, unwrap.
				types := input.Types()
				for _, t := range types.List() {
					result.Insert(t, pipeline.Name)
				}
			} else {
				// Not a user defined type, insert direct.
				result.Insert(inRef, pipeline.Name)
			}
		}
	}
	return result
}
