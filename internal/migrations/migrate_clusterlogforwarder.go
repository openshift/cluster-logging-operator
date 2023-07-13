package migrations

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
)

func MigrateClusterLogForwarderSpec(namespace, name string, spec loggingv1.ClusterLogForwarderSpec, logStore *loggingv1.LogStoreSpec, extras map[string]bool, logstoreSecretName string) (loggingv1.ClusterLogForwarderSpec, map[string]bool) {
	spec, extras = MigrateDefaultOutput(spec, logStore, extras, logstoreSecretName)
	if namespace == constants.OpenshiftNS && name == constants.SingletonName {
		spec.ServiceAccountName = constants.CollectorServiceAccountName
	}
	return spec, extras
}

// MigrateDefaultOutput adds the 'default' output spec to the list of outputs if it is not defined or
// selectively replaces it if it is.  It will apply OutputDefaults unless they are already defined.
func MigrateDefaultOutput(spec loggingv1.ClusterLogForwarderSpec, logStore *loggingv1.LogStoreSpec, extras map[string]bool, logstoreSecretName string) (loggingv1.ClusterLogForwarderSpec, map[string]bool) {
	// ClusterLogging without ClusterLogForwarder
	if len(spec.Pipelines) == 0 && len(spec.Inputs) == 0 && len(spec.Outputs) == 0 && spec.OutputDefaults == nil {
		if logStore != nil {
			log.V(2).Info("ClusterLogForwarder forwarding to default store")
			spec.Pipelines = []loggingv1.PipelineSpec{
				{
					InputRefs:  []string{loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure},
					OutputRefs: []string{loggingv1.OutputNameDefault},
				},
			}
			if logStore.Type == loggingv1.LogStoreTypeElasticsearch {
				spec.Outputs = []loggingv1.OutputSpec{NewDefaultOutput(nil, logstoreSecretName)}
				spec.Pipelines[0].Name = "default_pipeline_0_"
			}
		}
	}

	if logStore != nil && logStore.Type == loggingv1.LogStoreTypeLokiStack {
		var outputs []loggingv1.OutputSpec
		var pipelines []loggingv1.PipelineSpec
		outputs, pipelines, extras = lokistack.ProcessForwarderPipelines(logStore, constants.OpenshiftNS, spec, extras)

		spec.Outputs = outputs
		spec.Pipelines = pipelines

		return spec, extras
	}

	// Migrate ES ClusterLogForwarder
	routes := loggingv1.NewRoutes(spec.Pipelines)
	if _, ok := routes.ByOutput[loggingv1.OutputNameDefault]; ok {
		if logStore == nil {
			log.V(1).Info("ClusterLogForwarder references default logstore but one is not spec'd")
			return spec, extras
		} else {
			found := false
			defaultOutput := NewDefaultOutput(spec.OutputDefaults, logstoreSecretName)
			outputs := []loggingv1.OutputSpec{}
			for _, output := range spec.Outputs {
				if output.Name == loggingv1.OutputNameDefault {
					found = true
					if output.Type != loggingv1.OutputTypeElasticsearch {
						// append and continue so it will fail validation
						outputs = append(outputs, output)
						continue
					}
					if output.Elasticsearch != nil {
						defaultOutput.Elasticsearch = output.Elasticsearch
					}
					output = defaultOutput
					// We set this, so we know it was our default that was migrated
					extras[constants.MigrateDefaultOutput] = true
				}
				outputs = append(outputs, output)
			}
			if !found {
				// default output was never found so create it
				outputs = append(outputs, defaultOutput)
				extras[constants.MigrateDefaultOutput] = true
			}

			spec.Outputs = outputs
			return spec, extras
		}
	}
	return spec, extras
}

func NewDefaultOutput(defaults *loggingv1.OutputDefaults, logstoreSecretName string) loggingv1.OutputSpec {
	spec := loggingv1.OutputSpec{
		Name:   loggingv1.OutputNameDefault,
		Type:   loggingv1.OutputTypeElasticsearch,
		URL:    constants.LogStoreURL,
		Secret: &loggingv1.OutputSecretSpec{Name: logstoreSecretName},
	}
	if defaults != nil && defaults.Elasticsearch != nil {
		spec.Elasticsearch = &loggingv1.Elasticsearch{
			ElasticsearchStructuredSpec: *defaults.Elasticsearch,
		}
	}
	return spec
}
