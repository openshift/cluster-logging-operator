package passthrough

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func New(name string, inputs ...string) generator.Element {
	vrl := common.SrcPassThrough
	r := elements.Remap{
		ComponentID: name,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
	return r
}
