package api

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConvertLoggingToObservability converts a logging.ClusterLogging and/or logging.ClusterLogForwarder
// to observability.ClusterLogForwarder
func ConvertLoggingToObservability(k8sClient client.Client, loggingCL *logging.ClusterLogging, loggingCLF *logging.ClusterLogForwarder, outputSecrets map[string]*corev1.Secret) (*obs.ClusterLogForwarder, error) {
	// Legacy Workflow
	if loggingCL != nil && loggingCL.Name == constants.SingletonName && loggingCL.Namespace == constants.OpenshiftNS {
		// Create logcollector SA permissions for legacy case
		if err := CreateLogCollectorSAPermissions(k8sClient); err != nil {
			return nil, err
		}

		clToObsCLF := obsruntime.NewClusterLogForwarder(loggingCL.Namespace, loggingCL.Name, runtime.Initialize)
		// Legacy ClusterLogging only
		if loggingCLF == nil {
			obsClfSpec := convertLegacyClusterLogging(&loggingCL.Spec)
			clToObsCLF.Spec = *obsClfSpec
			return clToObsCLF, nil
		}
		// Legacy ClusterLogging + ClusterLogForwarder
		clToObsCLF.Spec = convertClusterLogForwarder(loggingCL, loggingCLF, outputSecrets, true)
		return clToObsCLF, nil
	}

	// custom clf
	obsCLF := obsruntime.NewClusterLogForwarder(loggingCLF.Namespace, loggingCLF.Name, runtime.Initialize)
	obsCLF.Spec = convertClusterLogForwarder(loggingCL, loggingCLF, outputSecrets, false)

	return obsCLF, nil
}

// convertLegacyClusterLogging generates the output and pipelines determined by the logstore type of the ClusterLoggingInstance
func convertLegacyClusterLogging(loggingCLSpec *logging.ClusterLoggingSpec) *obs.ClusterLogForwarderSpec {
	obsDefaultOut := *generateDefaultOutput(loggingCLSpec.LogStore)

	return &obs.ClusterLogForwarderSpec{
		ManagementState: obs.ManagementState(loggingCLSpec.ManagementState),
		ServiceAccount: obs.ServiceAccount{
			Name: constants.CollectorServiceAccountName,
		},
		Outputs: []obs.OutputSpec{
			obsDefaultOut,
		},
		Pipelines: []obs.PipelineSpec{
			{
				Name:       obsDefaultOut.Name + "-pipeline",
				InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeInfrastructure)},
				OutputRefs: []string{obsDefaultOut.Name},
			},
		},
		Collector: setCollectorResources(loggingCLSpec.Collection),
	}
}

func convertClusterLogForwarder(loggingCl *logging.ClusterLogging, loggingClf *logging.ClusterLogForwarder, secrets map[string]*corev1.Secret, isLegacyDeployment bool) obs.ClusterLogForwarderSpec {
	var logStoreSpec *logging.LogStoreSpec
	saName := constants.CollectorServiceAccountName

	if !isLegacyDeployment {
		saName = loggingClf.Spec.ServiceAccountName
	}

	// Set CLF service account
	obsClfSpec := &obs.ClusterLogForwarderSpec{
		ServiceAccount: obs.ServiceAccount{
			Name: saName,
		},
	}

	if loggingCl != nil {
		obsClfSpec.Collector = setCollectorResources(loggingCl.Spec.Collection)
		if loggingCl.Spec.LogStore != nil {
			logStoreSpec = loggingCl.Spec.LogStore
		}
	}

	obsClfSpec.Inputs = convertInputs(&loggingClf.Spec)
	obsClfSpec.Outputs = convertOutputs(&loggingClf.Spec, secrets)
	obsClfSpec.Filters = convertFilters(&loggingClf.Spec)

	obsPipelineSpec, filtersToAdd, needDefault := convertPipelines(logStoreSpec, &loggingClf.Spec)
	obsClfSpec.Pipelines = obsPipelineSpec
	// Add pipeline filters to clf.spec.filters
	obsClfSpec.Filters = append(obsClfSpec.Filters, filtersToAdd...)

	// Generate default output if referenced
	if needDefault && logStoreSpec != nil {
		obsClfSpec.Outputs = append(obsClfSpec.Outputs, *generateDefaultOutput(logStoreSpec))
	}

	return *obsClfSpec
}

func setCollectorResources(loggingCollSpec *logging.CollectionSpec) *obs.CollectorSpec {
	if loggingCollSpec == nil {
		return nil
	}

	obsCollSpecs := &obs.CollectorSpec{
		Resources:    loggingCollSpec.Resources,
		NodeSelector: loggingCollSpec.NodeSelector,
		Tolerations:  loggingCollSpec.Tolerations,
	}

	// If using spec.collection.logs, use that spec instead
	if loggingCollSpec.Logs != nil {
		obsCollSpecs = &obs.CollectorSpec{
			Resources:    loggingCollSpec.Logs.Resources,
			NodeSelector: loggingCollSpec.Logs.NodeSelector,
			Tolerations:  loggingCollSpec.Logs.Tolerations,
		}
	}

	return obsCollSpecs
}
