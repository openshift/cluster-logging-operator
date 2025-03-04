package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/utils/set"
)

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

// ReferenceOutput iterates through the list of pipelines to see if any reference the given output
func (pipeline Pipelines) ReferenceOutput(output obs.OutputSpec) bool {
	for _, i := range pipeline {
		if set.New(i.OutputRefs...).Has(output.Name) {
			return true
		}
	}
	return false
}
