package migrations

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func MigrateClusterLogForwarderSpec(spec logging.ClusterLogForwarderSpec) logging.ClusterLogForwarderSpec {
	spec.Outputs = MigrateDefaultOutput(spec)
	return spec
}

// MigrateDefaultOutput adds the 'default' output spec to the list of outputs if it is not defined or
// selectively replaces it if it is.  It will apply OutputDefaults unless they are already defined.
func MigrateDefaultOutput(spec logging.ClusterLogForwarderSpec) []logging.OutputSpec {

	refDefault := false
	for _, p := range spec.Pipelines {
		for _, ref := range p.OutputRefs {
			if ref == logging.OutputNameDefault {
				refDefault = true
				break
			}
		}
	}
	if !refDefault {
		return spec.Outputs
	}
	defaultReplaced := false
	defaultOutput := NewDefaultOutput(spec.OutputDefaults)
	outputs := []logging.OutputSpec{}
	for _, output := range spec.Outputs {
		if output.Name == logging.OutputNameDefault {
			if output.Elasticsearch != nil {
				defaultOutput.Elasticsearch = output.Elasticsearch
			}
			defaultReplaced = true
			output = defaultOutput
		}
		outputs = append(outputs, output)
	}
	if !defaultReplaced {
		outputs = append(outputs, defaultOutput)
	}
	return outputs
}

func NewDefaultOutput(defaults *logging.OutputDefaults) logging.OutputSpec {
	spec := logging.OutputSpec{
		Name:   logging.OutputNameDefault,
		Type:   logging.OutputTypeElasticsearch,
		URL:    constants.LogStoreURL,
		Secret: &logging.OutputSecretSpec{Name: constants.CollectorSecretName},
	}
	if defaults != nil && defaults.Elasticsearch != nil {
		spec.Elasticsearch = defaults.Elasticsearch
	}
	return spec
}
