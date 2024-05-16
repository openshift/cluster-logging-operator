package observability

import obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

// FilterMap returns a map of filter names to FilterSpec.
func FilterMap(spec obs.ClusterLogForwarderSpec) map[string]*obs.FilterSpec {
	m := map[string]*obs.FilterSpec{}
	for i := range spec.Filters {
		m[spec.Filters[i].Name] = &spec.Filters[i]
	}
	return m
}
