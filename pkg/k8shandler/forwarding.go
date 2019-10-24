package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
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

var (
	outputTypes = sets.NewString(string(logging.OutputTypeElasticsearch), string(logging.OutputTypeForward))
	sourceTypes = sets.NewString(string(logging.LogSourceTypeApp), string(logging.LogSourceTypeInfra))
)

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {

	switch clusterRequest.cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.cluster.Spec.Collection.Logs.Type)
	}

	clusterRequest.ForwardingRequest = clusterRequest.normalizeLogForwarding(clusterRequest.cluster.Namespace, clusterRequest.cluster)
	generator, err := forwarding.NewConfigGenerator(clusterRequest.cluster.Spec.Collection.Logs.Type)
	return generator.Generate(&clusterRequest.ForwardingRequest)
}

func (clusterRequest *ClusterLoggingRequest) normalizeLogForwarding(namespace string, cluster *logging.ClusterLogging) logging.ForwardingSpec {
	if cluster.Spec.LogStore != nil && cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {
		if cluster.Spec.Forwarding == nil || (!cluster.Spec.Forwarding.DisableDefaultForwarding && len(cluster.Spec.Forwarding.Pipelines) == 0) {
			cluster.Status.Forwarding = &logging.ForwardingStatus{
				LogSources: []string{string(logging.LogSourceTypeApp), string(logging.LogSourceTypeInfra)},
				Outputs: []logging.OutputStatus{
					logging.OutputStatus{
						Name:    internalOutputName,
						State:   logging.OutputStateAccepted,
						Message: "This is an operator generated output because forwarding is undefined and 'DisableDefaultForwarding' is false",
					},
				},
				Pipelines: []logging.PipelineStatus{
					logging.PipelineStatus{
						Name:    defaultAppPipelineName,
						State:   logging.PipelineStateAccepted,
						Message: "This is an operator generated pipeline because forwarding is undefined and 'DisableDefaultForwarding' is false",
					},
				},
			}
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
	logSources := sets.NewString()
	pipelineNames := sets.NewString()
	cluster.Status.Forwarding = &logging.ForwardingStatus{}
	var outputRefs sets.String
	outputRefs, normalized.Outputs = clusterRequest.gatherAndVerifyOutputRefs(cluster.Spec.Forwarding, cluster.Status.Forwarding)
	for i, pipeline := range cluster.Spec.Forwarding.Pipelines {
		status := logging.PipelineStatus{
			Name: pipeline.Name,
		}
		if pipeline.Name == "" {
			status.Name = fmt.Sprintf("pipeline[%d]", i)
			status.Reasons = append(status.Reasons, logging.PipelineStateReasonMissingName)
			status.State = logging.PipelineStateDropped
		}
		if pipeline.Name == defaultAppPipelineName || pipeline.Name == defaultInfraPipelineName {
			status.Name = fmt.Sprintf("pipeline[%d]", i)
			status.Reasons = append(status.Reasons, logging.PipelineStateReasonReservedNameConflict)
			status.State = logging.PipelineStateDropped
		}
		if pipelineNames.Has(pipeline.Name) {
			status.Name = fmt.Sprintf("pipeline[%d]", i)
			status.Reasons = append(status.Reasons, logging.PipelineStateReasonNonUniqueName)
			status.State = logging.PipelineStateDropped
		}
		if string(pipeline.SourceType) == "" {
			status.Reasons = append(status.Reasons, logging.PipelineStateReasonMissingSource)
			status.State = logging.PipelineStateDropped
		}
		if !sourceTypes.Has(string(pipeline.SourceType)) {
			status.Reasons = append(status.Reasons, logging.PipelineStateReasonUnrecognizedSource)
			status.State = logging.PipelineStateDropped
		}
		if status.State != logging.PipelineStateDropped {
			newPipeline := logging.PipelineSpec{
				Name:       pipeline.Name,
				SourceType: pipeline.SourceType,
			}
			for _, output := range pipeline.OutputRefs {
				if outputRefs.Has(output) {
					newPipeline.OutputRefs = append(newPipeline.OutputRefs, output)
				} else {
					logger.Warnf("OutputRef %q for forwarding pipeline %q was not defined", output, pipeline.Name)
					status.Reasons = append(status.Reasons, logging.PipelineStateReasonUnrecognizedOutput)
				}
			}
			if len(newPipeline.OutputRefs) > 0 {
				pipelineNames.Insert(pipeline.Name)
				logSources.Insert(string(pipeline.SourceType))
				normalized.Pipelines = append(normalized.Pipelines, newPipeline)
				status.State = logging.PipelineStateAccepted
				if len(newPipeline.OutputRefs) != len(pipeline.OutputRefs) {
					status.State = logging.PipelineStateDegraded
					status.Reasons = append(status.Reasons, logging.PipelineStateReasonMissingOutputs)
				}
			} else {
				logger.Warnf("Dropping forwarding pipeline %q as its ouptutRefs have no corresponding outputs", pipeline.Name)
				status.State = logging.PipelineStateDropped
				status.Reasons = append(status.Reasons, logging.PipelineStateReasonMissingOutputs)
			}
		}

		cluster.Status.Forwarding.Pipelines = append(cluster.Status.Forwarding.Pipelines, status)
	}
	cluster.Status.Forwarding.LogSources = logSources.List()

	return normalized
}

func (clusterRequest *ClusterLoggingRequest) gatherAndVerifyOutputRefs(spec *logging.ForwardingSpec, status *logging.ForwardingStatus) (sets.String, []logging.OutputSpec) {
	refs := sets.NewString()
	outputs := []logging.OutputSpec{}
	for i, output := range spec.Outputs {
		outStatus := logging.OutputStatus{
			Name:  output.Name,
			State: logging.OutputStateDropped,
		}
		if output.Name == "" {
			outStatus.Name = fmt.Sprintf("output[%d]", i)
			outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReasonMissingName)
		}
		if output.Name == internalOutputName {
			outStatus.Name = fmt.Sprintf("output[%d]", i)
			outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReservedNameConflict)
		}
		if refs.Has(output.Name) {
			outStatus.Name = fmt.Sprintf("output[%d]", i)
			outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateNonUniqueName)
			outStatus.Message = outStatus.Message + "The output name is not unique among all defined outputs."
		}
		if string(output.Type) == "" {
			outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReasonMissingType)
		}
		if !outputTypes.Has(string(output.Type)) {
			outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReasonUnrecognizedType)
		}
		if output.Endpoint == "" {
			outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReasonMissingEndpoint)
		}
		if output.Secret != nil {
			if output.Secret.Name == "" {
				outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReasonMissingSecretName)
			} else {
				_, err := clusterRequest.GetSecret(output.Secret.Name)
				if errors.IsNotFound(err) {
					outStatus.Reasons = append(outStatus.Reasons, logging.OutputStateReasonSecretDoesNotExist)
				}
			}
		}

		if len(outStatus.Reasons) == 0 {
			outStatus.State = logging.OutputStateAccepted
			refs.Insert(output.Name)
			outputs = append(outputs, output)
		}
		logger.Debugf("Status of output evaluation: %v", outStatus)
		status.Outputs = append(status.Outputs, outStatus)

	}
	return refs, outputs
}
