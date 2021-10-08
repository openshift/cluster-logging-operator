package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/kafka"
	transform "github.com/openshift/cluster-logging-operator/internal/generator/vector/transform"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func getInputPipelines(transformers []transform.Transform, inputRefs sets.String) sets.String {
	inputSet := sets.NewString()
	for _, element := range transformers {
		if inputRefs.Has(element.SourceType()) {
			inputSet.Insert(element.ComponentID())
		}
	}
	return inputSet
}

func getOutputMap(clfspec *logging.ClusterLogForwarderSpec, transformers []transform.Transform) map[string]sets.String {
	outputTypeToInputPipelineMap := map[string]sets.String{}

	output_input_map := logging.NewRoutes(clfspec.Pipelines).ByOutput
	for outputRef := range output_input_map {
		inputRefSet := output_input_map[outputRef]
		inputSet := getInputPipelines(transformers, inputRefSet)
		outputTypeToInputPipelineMap[outputRef] = inputSet

	}
	return outputTypeToInputPipelineMap
}

func Outputs(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op Options, transformers []transform.Transform) []Element {
	outputs := []Element{
		Comment("Ship logs to specific outputs"),
	}
	var bufspec *logging.FluentdBufferSpec = nil
	if clspec != nil &&
		clspec.Forwarder != nil &&
		clspec.Forwarder.Fluentd != nil &&
		clspec.Forwarder.Fluentd.Buffer != nil {
		bufspec = clspec.Forwarder.Fluentd.Buffer
	}
	outputTypeToInputPipelineMap := getOutputMap(clfspec, transformers)

	for _, o := range clfspec.Outputs {
		secret := secrets[o.Name]
		inputPipelines := outputTypeToInputPipelineMap[o.Name].UnsortedList()
		switch o.Type {
		case logging.OutputTypeKafka:
			outputs = MergeElements(outputs, kafka.Conf(bufspec, secret, o, op, inputPipelines))
		}
	}

	return outputs
}
