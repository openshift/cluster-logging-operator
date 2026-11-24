package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

// CollectorImplementation returns the collector implementation the forwarder
// requests via the collector implementation annotation, defaulting to vector
func CollectorImplementation(forwarder obs.ClusterLogForwarder) string {
	if impl, found := forwarder.Annotations[constants.AnnotationCollectorImplementation]; found && impl == constants.ComponentNameOtelc {
		return impl
	}
	return constants.VectorName
}

// IsOtelc returns true when the forwarder deploys the OpenTelemetry Collector
func IsOtelc(forwarder obs.ClusterLogForwarder) bool {
	return CollectorImplementation(forwarder) == constants.ComponentNameOtelc
}
