package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/transform"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Section {

	sources := Sources(clfspec, op)
	transformersList := InputsToPipeline(clfspec, op, sources)
	outputList := Outputs(clspec, secrets, clfspec, op, transformersList)
	elementList := merge(sources, transformersList, outputList)

	sectionList := make([]generator.Section, 0)
	sectionList = append(sectionList, generator.Section{Elements: elementList})
	return sectionList
}

func merge(sources []source.LogSource, transformers []transform.Transform, outputs []generator.Element) []generator.Element {
	merged := make([]generator.Element, 0)
	for _, source := range sources {
		element := generator.Element(source)
		merged = append(merged, element)
	}

	for _, transformer := range transformers {
		element := generator.Element(transformer)
		merged = append(merged, element)
	}

	for _, output := range outputs {
		merged = append(merged, output)
	}

	return merged
}
