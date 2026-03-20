package adapters

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

// PipelineFilter is a dedicated instance of the CLF filter for the pipeline
type PipelineFilter struct {
	ids     []string
	Next    []helpers.InputComponent
	Factory func(inputs ...string) types.Transform
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

func NewPipelineFilter(pipelineName, filterRef string, spec InternalFilterSpec) *PipelineFilter {
	ids := []string{helpers.MakePipelineID(pipelineName, filterRef)}
	return &PipelineFilter{
		ids:     ids,
		Factory: spec.Factory,
	}
}

// Transform creates an instance of a transform based upon the instance of a filter referenced by a pipeline
func (pf *PipelineFilter) Transform() types.Transform {
	inputs := []string{}
	for _, n := range pf.Next {
		if n != nil {
			inputs = append(inputs, n.InputIDs()...)
		}
	}
	sort.Strings(inputs)
	return pf.Factory(inputs...)
}
