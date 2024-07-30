package filters

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapPruneFilter(loggingPruneSpec logging.PruneFilterSpec) *obs.PruneFilterSpec {
	spec := &obs.PruneFilterSpec{}
	for _, in := range loggingPruneSpec.In {
		spec.In = append(spec.In, obs.FieldPath(in))
	}
	for _, notIn := range loggingPruneSpec.NotIn {
		spec.NotIn = append(spec.NotIn, obs.FieldPath(notIn))
	}

	return spec
}
