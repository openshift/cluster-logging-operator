package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func InputsToPipelines(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	el := []generator.Element{}
	for _, p := range spec.Pipelines {
		r := Remap{
			ComponentID: p.Name,
			Inputs:      helpers.MakeInputs(p.InputRefs...),
			VRL:         PassThrough,
		}
		el = append(el, r)

	}
	return el
}
