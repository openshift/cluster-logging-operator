package telemetry

import (
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
)

func gatherForwarderInfo(forwarder *observabilityv1.ClusterLogForwarder) (pipelines uint, inputs, outputs []string) {
	inputs = makeZeroStrings(len(forwarderInputTypes))
	outputs = makeZeroStrings(len(forwarderOutputTypes))

	outputMap := internalobs.Outputs(forwarder.Spec.Outputs).Map()

	activeInputNames := map[string]bool{}
	activeOutputTypes := map[string]bool{}
	for _, pipeline := range forwarder.Spec.Pipelines {
		for _, inputName := range pipeline.InputRefs {
			activeInputNames[inputName] = true
		}

		for _, outputName := range pipeline.OutputRefs {
			output, found := outputMap[outputName]
			if found {
				activeOutputTypes[string(output.Type)] = true
			}
		}
	}

	pipelines = uint(len(forwarder.Spec.Pipelines))

	for i, v := range forwarderInputTypes {
		if activeInputNames[v] {
			inputs[i] = boolYes
		}
	}

	for i, v := range forwarderOutputTypes {
		if activeOutputTypes[v] {
			outputs[i] = boolYes
		}
	}

	return pipelines, inputs, outputs
}
