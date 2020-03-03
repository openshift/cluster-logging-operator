package k8shandler

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logforward "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

const (
	internalOutputName       = "clo-default-output-es"
	collectorSecretName      = "fluentd"
	defaultAppPipelineName   = "clo-default-app-pipeline"
	defaultInfraPipelineName = "clo-default-infra-pipeline"

	//ForwardingAnnotation  Annotate CL instance with a value of "enabled"
	ForwardingAnnotation = "clusterlogging.openshift.io/logforwardingtechpreview"
)

var (
	outputTypes = sets.NewString(string(logforward.OutputTypeElasticsearch), string(logforward.OutputTypeForward), string(logforward.OutputTypeSyslog))
	sourceTypes = sets.NewString(string(logforward.LogSourceTypeApp), string(logforward.LogSourceTypeInfra), string(logforward.LogSourceTypeAudit))
)

func isForwardingEnabled(cluster *logging.ClusterLogging) bool {
	if value, _ := utils.GetAnnotation(ForwardingAnnotation, cluster.ObjectMeta); value == "enabled" {
		return true
	}
	return false
}

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {

	switch clusterRequest.cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.cluster.Spec.Collection.Logs.Type)
	}

	clusterRequest.ForwardingSpec = clusterRequest.normalizeLogForwarding(clusterRequest.cluster.Namespace, clusterRequest.cluster)
	generator, err := forwarding.NewConfigGenerator(clusterRequest.cluster.Spec.Collection.Logs.Type, clusterRequest.includeLegacyForwardConfig(), clusterRequest.includeLegacySyslogConfig())
	return generator.Generate(&clusterRequest.ForwardingSpec)
}

func (clusterRequest *ClusterLoggingRequest) normalizeLogForwarding(namespace string, cluster *logging.ClusterLogging) logforward.ForwardingSpec {
	logger.Debugf("Normalizing logforwarding from request: %v", clusterRequest)
	logger.Debugf("ForwardingRequest: %v", clusterRequest.ForwardingRequest)
	logger.Debugf("ForwardingSpec: %v", clusterRequest.ForwardingSpec)
	if cluster.Spec.LogStore != nil && cluster.Spec.LogStore.Type == logging.LogStoreTypeElasticsearch {
		if !clusterRequest.ForwardingSpec.DisableDefaultForwarding && len(clusterRequest.ForwardingSpec.Pipelines) == 0 {
			logger.Debug("Configuring logforwarding to utilize the operator managed logstore")
			if clusterRequest.ForwardingRequest.Status == nil {
				clusterRequest.ForwardingRequest.Status = logforward.NewForwardingStatus(logforward.LogForwardingStateAccepted, logforward.LogForwardingReasonName, "")
			}
			clusterRequest.ForwardingRequest.Status.LogSources = []logforward.LogSourceType{logforward.LogSourceTypeApp, logforward.LogSourceTypeInfra}
			clusterRequest.ForwardingRequest.Status.Outputs = []logforward.OutputStatus{
				logforward.NewOutputStatus(
					internalOutputName,
					logforward.OutputStateAccepted,
					logforward.OutputStateReasonConditionsMet,
					"This is an operator generated output because forwarding is undefined and 'DisableDefaultForwarding' is false",
				),
			}
			clusterRequest.ForwardingRequest.Status.Pipelines = []logforward.PipelineStatus{
				logforward.NewPipelineStatus(
					defaultAppPipelineName,
					logforward.PipelineStateAccepted,
					logforward.PipelineStateReasonConditionsMet,
					"This is an operator generated pipeline because forwarding is undefined and 'DisableDefaultForwarding' is false",
				),
			}
			return logforward.ForwardingSpec{
				Outputs: []logforward.OutputSpec{
					logforward.OutputSpec{
						Name:     internalOutputName,
						Type:     logforward.OutputTypeElasticsearch,
						Endpoint: constants.LogStoreService,
						Secret: &logforward.OutputSecretSpec{
							Name: collectorSecretName,
						},
					},
				},
				Pipelines: []logforward.PipelineSpec{
					logforward.PipelineSpec{
						Name:       defaultAppPipelineName,
						SourceType: logforward.LogSourceTypeApp,
						OutputRefs: []string{internalOutputName},
					},
					logforward.PipelineSpec{
						Name:       defaultInfraPipelineName,
						SourceType: logforward.LogSourceTypeInfra,
						OutputRefs: []string{internalOutputName},
					},
				},
			}
		}
	}
	normalized := logforward.ForwardingSpec{}
	if !isForwardingEnabled(cluster) || clusterRequest.ForwardingRequest == nil {
		return normalized
	}
	logSources := logforward.NewLogSourceTypeSet()
	pipelineNames := sets.NewString()
	clusterRequest.ForwardingRequest.Status = &logforward.ForwardingStatus{}
	var outputRefs sets.String
	outputRefs, normalized.Outputs = clusterRequest.gatherAndVerifyOutputRefs(&clusterRequest.ForwardingSpec, clusterRequest.ForwardingRequest.Status)
	for i, pipeline := range clusterRequest.ForwardingSpec.Pipelines {
		status := logforward.NewPipelineStatusNamed(pipeline.Name)
		if pipeline.Name == "" {
			status.Name = fmt.Sprintf("pipeline[%d]", i)
			status.AddCondition(logforward.PipelineConditionTypeName, logforward.PipelineConditionReasonMissingName, "")
		}
		if pipeline.Name == defaultAppPipelineName || pipeline.Name == defaultInfraPipelineName {
			status.Name = fmt.Sprintf("pipeline[%d]", i)
			status.AddCondition(logforward.PipelineConditionTypeName, logforward.PipelineConditionReasonReservedNameConflict, "")
		}
		if pipelineNames.Has(pipeline.Name) {
			status.Name = fmt.Sprintf("pipeline[%d]", i)
			status.AddCondition(logforward.PipelineConditionTypeName, logforward.PipelineConditionReasonUniqueName, "")
		}
		if string(pipeline.SourceType) == "" {
			status.AddCondition(logforward.PipelineConditionTypeSourceType, logforward.PipelineConditionReasonMissingSource, "")
		}
		if !sourceTypes.Has(string(pipeline.SourceType)) {
			status.AddCondition(logforward.PipelineConditionTypeSourceType, logforward.PipelineConditionReasonUnrecognizedSourceType, "")
		}
		if len(status.Conditions) > 0 {
			status.State = logforward.PipelineStateDropped
		}
		if status.State != logforward.PipelineStateDropped {
			newPipeline := logforward.PipelineSpec{
				Name:       pipeline.Name,
				SourceType: pipeline.SourceType,
			}
			for _, output := range pipeline.OutputRefs {
				if outputRefs.Has(output) {
					newPipeline.OutputRefs = append(newPipeline.OutputRefs, output)
				} else {
					logger.Warnf("OutputRef %q for forwarding pipeline %q was not defined", output, pipeline.Name)
					status.AddCondition(logforward.PipelineConditionTypeOutputRef, logforward.PipelineConditionReasonUnrecognizedOutputRef, "")
				}
			}
			if len(newPipeline.OutputRefs) > 0 {
				pipelineNames.Insert(pipeline.Name)
				logSources.Insert(pipeline.SourceType)
				normalized.Pipelines = append(normalized.Pipelines, newPipeline)
				status.State = logforward.PipelineStateAccepted
				if len(newPipeline.OutputRefs) != len(pipeline.OutputRefs) {
					status.State = logforward.PipelineStateDegraded
					status.AddCondition(logforward.PipelineConditionTypeOutputRef, logforward.PipelineConditionReasonMissingOutputs, "")
				}
			} else {
				logger.Warnf("Dropping forwarding pipeline %q as its ouptutRefs have no corresponding outputs", pipeline.Name)
				status.State = logforward.PipelineStateDropped
				status.AddCondition(logforward.PipelineConditionTypeOutputRef, logforward.PipelineConditionReasonMissingOutputs, "")
			}
		}

		clusterRequest.ForwardingRequest.Status.Pipelines = append(clusterRequest.ForwardingRequest.Status.Pipelines, status)
	}
	clusterRequest.ForwardingRequest.Status.LogSources = logSources.List()

	return normalized
}

func (clusterRequest *ClusterLoggingRequest) gatherAndVerifyOutputRefs(spec *logforward.ForwardingSpec, status *logforward.ForwardingStatus) (sets.String, []logforward.OutputSpec) {
	refs := sets.NewString()
	outputs := []logforward.OutputSpec{}
	for i, output := range spec.Outputs {
		outStatus := logforward.NewOutputStatusNamed(output.Name)
		outStatus.State = logforward.OutputStateDropped
		if output.Name == "" {
			outStatus.Name = fmt.Sprintf("output[%d]", i)
			outStatus.AddCondition(logforward.OutputConditionTypeName, logforward.OutputConditionReasonMissingName, "")
		}
		if output.Name == internalOutputName {
			outStatus.Name = fmt.Sprintf("output[%d]", i)
			outStatus.AddCondition(logforward.OutputConditionTypeName, logforward.OutputConditionReasonReservedNameConflict, "")
		}
		if refs.Has(output.Name) {
			outStatus.Name = fmt.Sprintf("output[%d]", i)
			outStatus.AddCondition(logforward.OutputConditionTypeName, logforward.OutputConditionReasonNonUniqueName, "The output name is not unique among all defined outputs.")
		}
		if string(output.Type) == "" {
			outStatus.AddCondition(logforward.OutputConditionTypeType, logforward.OutputConditionReasonMissingType, "")
		}
		if !outputTypes.Has(string(output.Type)) {
			outStatus.AddCondition(logforward.OutputConditionTypeType, logforward.OutputConditionReasonUnrecognizedType, "")
		}
		if output.Endpoint == "" {
			outStatus.AddCondition(logforward.OutputConditionTypeEndpoint, logforward.OutputConditionReasonMissingEndpoint, "")
		}
		if output.Secret != nil {
			if output.Secret.Name == "" {
				outStatus.AddCondition(logforward.OutputConditionTypeSecret, logforward.OutputConditionReasonMissingSecretName, "")
			} else {
				secret, err := clusterRequest.GetSecret(output.Secret.Name)
				if errors.IsNotFound(err) {
					outStatus.AddCondition(logforward.OutputConditionTypeSecret, logforward.OutputConditionReasonSecretDoesNotExist, "")
				}
				verifyOutputSecret(output, secret, &outStatus)
			}
		}

		if len(outStatus.Conditions) == 0 {
			outStatus.State = logforward.OutputStateAccepted
			refs.Insert(output.Name)
			outputs = append(outputs, output)
		}
		logger.Debugf("Status of output evaluation: %v", outStatus)
		status.Outputs = append(status.Outputs, outStatus)

	}
	return refs, outputs
}

func verifyOutputSecret(output logforward.OutputSpec, secret *core.Secret, status *logforward.OutputStatus) {
	if output.Type != logforward.OutputTypeForward || secret == nil {
		return
	}
	if _, exists := secret.Data["shared_key"]; !exists {
		status.AddCondition(logforward.OutputConditionTypeSecret, logforward.OutputConditionReasonSecretMissingSharedKey, "")
	}
}
