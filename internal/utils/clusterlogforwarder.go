package utils

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

// OutputMap returns a map of names to outputs.
func OutputMap(spec *obs.ClusterLogForwarderSpec) map[string]*obs.OutputSpec {
	m := map[string]*obs.OutputSpec{}
	for i := range spec.Outputs {
		m[spec.Outputs[i].Name] = &spec.Outputs[i]
	}
	return m
}
