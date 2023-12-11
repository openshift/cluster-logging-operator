package telemetry

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

// updateDefaultInfo is used to update the information about the "default ClusterLogForwarder" used by the
// telemetry metrics.
//
// The default instance (namespace: openshift-logging, name: instance) is treated differently from the other CLF
// instances, because in openshift-logging just having a ClusterLogging instance without an associated
// ClusterLogForwarder is a valid configuration. So the "default ClusterLogForwarder instance" might not even exist
// or, even when it exists, might not contain all the information.
//
// Although this function is called for each ClusterLogForwarder instance during reconciliation, it only saves
// the values of the default instance. The metrics for all the other ClusterLogForwarder instances are generated
// directly from the resources.
func (t *telemetryCollector) updateDefaultInfo(forwarder *loggingv1.ClusterLogForwarder) {
	if forwarder.Namespace != constants.OpenshiftNS || forwarder.Name != constants.SingletonName {
		return
	}

	pipelines, inputs, outputs := gatherForwarderInfo(forwarder)
	t.defaultCLFInfo.NumPipelines = pipelines
	t.defaultCLFInfo.Inputs = inputs
	t.defaultCLFInfo.Outputs = outputs
}

func gatherForwarderInfo(forwarder *loggingv1.ClusterLogForwarder) (pipelines uint, inputs, outputs []string) {
	inputs = makeZeroStrings(len(forwarderInputTypes))
	outputs = makeZeroStrings(len(forwarderOutputTypes))

	outputMap := forwarder.Spec.OutputMap()

	activeInputNames := map[string]bool{}
	activeOutputTypes := map[string]bool{}
	for _, pipeline := range forwarder.Spec.Pipelines {
		for _, inputName := range pipeline.InputRefs {
			activeInputNames[inputName] = true
		}

		for _, outputName := range pipeline.OutputRefs {
			output, found := outputMap[outputName]
			if found {
				activeOutputTypes[output.Type] = true
			}
		}
	}

	pipelines = uint(len(forwarder.Spec.Pipelines))

	for i, v := range forwarderInputTypes {
		if activeInputNames[v] {
			inputs[i] = "1"
		}
	}

	for i, v := range forwarderOutputTypes {
		if activeOutputTypes[v] {
			outputs[i] = "1"
		}
	}

	return pipelines, inputs, outputs
}
