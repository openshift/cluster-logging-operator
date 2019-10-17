package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
)

const (
	internalOutputName       = "clo-default-output-es"
	collectorSecretName      = "fluentd"
	logStoreService          = "elasticsearch.openshift-logging.svc:9200"
	defaultAppPipelineName   = "clo-default-app-pipeline"
	defaultInfraPipelineName = "clo-default-infra-pipeline"
	secureForwardConfHash    = "8163d9a59a20ada8ab58c2535a3a4924"
)

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {

	switch clusterRequest.cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.cluster.Spec.Collection.Logs.Type)
	}

	clusterRequest.ForwardingRequest = normalizeLogForwarding(clusterRequest.cluster.Namespace, clusterRequest.cluster)
	generator, err := forwarding.NewConfigGenerator(clusterRequest.cluster.Spec.Collection.Logs.Type)
	return generator.Generate(&clusterRequest.ForwardingRequest)
}

func normalizeLogForwarding(namespace string, cluster *logging.ClusterLogging) logging.ForwardingSpec {
	if cluster.Spec.LogStore != nil && cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {
		if cluster.Spec.Forwarding == nil || (!cluster.Spec.Forwarding.DisableDefaultForwarding && len(cluster.Spec.Forwarding.Pipelines) == 0) {
			return logging.ForwardingSpec{
				Outputs: []logging.OutputSpec{
					logging.OutputSpec{
						Name:     internalOutputName,
						Type:     logging.OutputTypeElasticsearch,
						Endpoint: logStoreService,
						Secret: &logging.OutputSecretSpec{
							Name: collectorSecretName,
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					logging.PipelineSpec{
						Name:       defaultAppPipelineName,
						SourceType: logging.LogSourceTypeApp,
						OutputRefs: []string{internalOutputName},
					},
					logging.PipelineSpec{
						Name:       defaultInfraPipelineName,
						SourceType: logging.LogSourceTypeInfra,
						OutputRefs: []string{internalOutputName},
					},
				},
			}
		}
	}
	normalized := logging.ForwardingSpec{}
	if cluster.Spec.Forwarding == nil {
		return normalized
	}
	normalized.Outputs = cluster.Spec.Forwarding.Outputs
	outputRefs := gatherOutputRefs(cluster.Spec.Forwarding)
	for _, pipeline := range cluster.Spec.Forwarding.Pipelines {
		newPipeline := logging.PipelineSpec{
			Name:       pipeline.Name,
			SourceType: pipeline.SourceType,
		}
		for _, output := range pipeline.OutputRefs {
			if outputRefs.Has(output) {
				newPipeline.OutputRefs = append(newPipeline.OutputRefs, output)
			} else {
				logger.Warnf("OutputRef %q for forwarding pipeline %q was not defined", output, pipeline.Name)
			}
		}
		if len(newPipeline.OutputRefs) > 0 {
			normalized.Pipelines = append(normalized.Pipelines, newPipeline)
		} else {
			logger.Warnf("Dropping forwarding pipeline %q as its ouptutRefs have no corresponding outputs", pipeline.Name)
		}
	}
	return normalized
}

func gatherOutputRefs(spec *logging.ForwardingSpec) sets.String {
	refs := sets.NewString()
	for _, output := range spec.Outputs {
		refs.Insert(output.Name)
	}
	return refs
}
