package generator

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const (
	IncludeLegacyForwardConfig      = "includeLegacyForwardConfig"
	IncludeLegacySyslogConfig       = "includeLegacySyslogConfig"
	UseOldRemoteSyslogPlugin        = "useOldRemoteSyslogPlugin"
	LegacySecureforward             = "_LEGACY_SECUREFORWARD"
	LegacySyslog                    = "_LEGACY_SYSLOG"
	LoggingNamespace                = "openshift-logging"
	FluentdCollectorPodNamePrefix   = "fluentd"
	FluentBitCollectorPodNamePrefix = "fluent-bit"
	VectorCollectorPodNamePrefix    = "vector"
	ESLogStorePodNamePrefix         = "elasticsearch"
	LokiLogStorePodNamePrefix       = "loki"
	VisualizationPodNamePrefix      = "kibana"
)

//GatherSources collects the set of unique source types and namespaces
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
		}
	}
	return types
}

func InputsToPipelines(fwdspec *logging.ClusterLogForwarderSpec) logging.RouteMap {
	result := logging.RouteMap{}
	inputs := fwdspec.InputMap()
	for _, pipeline := range fwdspec.Pipelines {
		for _, inRef := range pipeline.InputRefs {
			if input, ok := inputs[inRef]; ok {
				// User defined input spec, unwrap.
				for t := range input.Types() {
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

func AddLegacySources(types sets.String, op Options) sets.String {
	if IsIncludeLegacyForwardConfig(op) {
		types.Insert(logging.InputNameApplication)
		types.Insert(logging.InputNameInfrastructure)
	}
	if IsIncludeLegacySyslogConfig(op) {
		types.Insert(logging.InputNameApplication)
		types.Insert(logging.InputNameInfrastructure)
		types.Insert(logging.InputNameAudit)
	}
	return types
}

func IsIncludeLegacyForwardConfig(op Options) bool {
	_, ok := op[IncludeLegacyForwardConfig]
	return ok
}

func IsIncludeLegacySyslogConfig(op Options) bool {
	_, ok := op[IncludeLegacySyslogConfig]
	return ok
}

func CollectorLogsPath() string {
	return fmt.Sprintf("/var/log/containers/%%s-*_%s_*.log", LoggingNamespace)
}

func LogStoreLogsPath() string {
	return fmt.Sprintf("/var/log/containers/%%s-*_%s_*.log", LoggingNamespace)
}

func VisualizationLogsPath() string {
	return fmt.Sprintf("/var/log/containers/%s-*_%s_*.log", VisualizationPodNamePrefix, LoggingNamespace)
}
