package clusterlogforwarder

import (
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

// validateJsonParsingToElasticsearch validates that when a pipeline that includes an
// Elasticsearch output type enables JSON parsing that it defines structuredTypeKey
// ref: https://issues.redhat.com/browse/LOG-2759
func validateJsonParsingToElasticsearch(clf loggingv1.ClusterLogForwarder) error {
	outputs := clf.Spec.OutputMap()

	for _, pipeline := range clf.Spec.Pipelines {
		if pipeline.Parse == "json" {
			for _, name := range pipeline.OutputRefs {
				if output := outputs[name]; output != nil && output.Type == loggingv1.OutputTypeElasticsearch {
					switch {
					case output.Elasticsearch != nil && (output.Elasticsearch.StructuredTypeName != "" || output.Elasticsearch.StructuredTypeKey != ""):
						continue
					case clf.Spec.OutputDefaults != nil && clf.Spec.OutputDefaults.Elasticsearch != nil && (clf.Spec.OutputDefaults.Elasticsearch.StructuredTypeName != "" || clf.Spec.OutputDefaults.Elasticsearch.StructuredTypeKey != ""):
						continue
					default:
						return fmt.Errorf("structuredTypeKey or structuredTypeName must be defined for Elasticsearch output named %q when JSON parsing is enabled on pipeline %q that references it", name, pipeline.Name)
					}
				}
			}
		}
	}
	return nil
}
