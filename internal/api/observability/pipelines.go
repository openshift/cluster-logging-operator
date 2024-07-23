package observability

import obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

type Pipelines []obs.PipelineSpec

// Names returns a slice of pipeline names
func (pipeline Pipelines) Names() (names []string) {
	for _, p := range pipeline {
		names = append(names, p.Name)
	}
	return names
}

// Map returns a map of pipeline names to pipeline spec
func (pipeline Pipelines) Map() map[string]obs.PipelineSpec {
	m := map[string]obs.PipelineSpec{}
	for _, p := range pipeline {
		m[p.Name] = p
	}
	return m
}
