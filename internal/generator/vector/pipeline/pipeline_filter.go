package pipeline

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

// PipelineFilter is an adapter between CLF pipeline filter instance and config generation
type PipelineFilter struct {
	pipeline obs.PipelineSpec
	ids      []string
	Next     []helpers.InputComponent
	vrl      string
	// Distinguish between a Remap or Filter element
	isFilterElement bool

	//transformFactory is a function that takes input IDs and returns a transform
	transformFactory func(...string) framework.Element
}

func (pf *PipelineFilter) ID() string {
	return pf.ids[0]
}
func (pf *PipelineFilter) InputIDs() []string {
	return pf.ids
}

func (pf *PipelineFilter) AddInputFrom(n helpers.InputComponent) {
	pf.Next = append(pf.Next, n)
}

func NewPipelineFilter(pipelineName, filterRef string, spec filter.InternalFilterSpec, pipeline obs.PipelineSpec) *PipelineFilter {
	ids := []string{helpers.MakePipelineID(pipelineName, filterRef)}
	if spec.SuppliesTransform {
		return &PipelineFilter{
			ids: ids,
			transformFactory: func(inputs ...string) framework.Element {
				return spec.TranformFactory(ids[0], inputs...)
			},
		}
	}

	if vrl, err := spec.RemapFilter.VRL(); err != nil {
		log.Error(err, "bad filter", "filterRef", filterRef, "spec.type", spec.Type, "spec.Name", spec.Name)
		return nil
	} else {
		return &PipelineFilter{
			pipeline: pipeline,
			ids:      ids,
			vrl:      vrl,
			isFilterElement: func() bool {
				return spec.Type == obs.FilterTypeDrop
			}(),
		}
	}
}

func (o *PipelineFilter) Element() framework.Element {
	inputs := []string{}
	for _, n := range o.Next {
		if n != nil {
			inputs = append(inputs, n.InputIDs()...)
		}
	}
	if o.transformFactory != nil {
		return o.transformFactory(inputs...)
	}

	if o.isFilterElement {
		return elements.Filter{
			ComponentID: o.ids[0],
			Inputs:      helpers.MakeInputs(inputs...),
			Condition:   o.vrl,
		}
	}
	return elements.Remap{
		ComponentID: o.ids[0],
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         o.vrl,
	}
}
