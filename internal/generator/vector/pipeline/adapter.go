package pipeline

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"os"
	"strconv"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// Pipeline is an adapter between logging API and config generation
type Pipeline struct {
	obs.PipelineSpec
	index      int
	filterMap  map[string]filter.InternalFilterSpec
	Filters    []*PipelineFilter
	inputSpecs []obs.InputSpec
}

func (o *Pipeline) Elements() []framework.Element {
	elements := []framework.Element{}
	for _, pf := range o.Filters {
		elements = append(elements, pf.Element())
	}
	return elements
}

func NewPipeline(index int, p obs.PipelineSpec, inputs map[string]helpers.InputComponent, outputs map[string]*output.Output, filters map[string]*filter.InternalFilterSpec, inputSpecs []obs.InputSpec) *Pipeline {
	pipeline := &Pipeline{
		PipelineSpec: p,
		index:        index,
		filterMap:    map[string]filter.InternalFilterSpec{},
		inputSpecs:   []obs.InputSpec{},
	}
	for _, is := range inputSpecs {
		for _, ref := range p.InputRefs {
			if is.Name == ref {
				pipeline.inputSpecs = append(pipeline.inputSpecs, is)
			}
		}
	}
	for name, f := range filters {
		pipeline.filterMap[name] = *f
	}
	addPostFilters(pipeline)

	for i, filterName := range pipeline.FilterRefs {
		pipeline.initFilter(i, filterName)
	}

	if len(pipeline.FilterRefs) > 0 {
		if len(pipeline.Filters) == 0 {
			log.V(0).Info("Runtime error in pipelineAdapter while processing filters.  Filter spec'd but not constructed", "filterRefs", pipeline.FilterRefs)
			os.Exit(0)
		}
		first := pipeline.Filters[0]
		for _, inputRefs := range pipeline.InputRefs {
			first.AddInputFrom(inputs[inputRefs])
		}

		last := pipeline.Filters[len(pipeline.FilterRefs)-1]
		for _, name := range pipeline.OutputRefs {
			outputs[name].AddInputFrom(last)
		}
	} else {
		for _, outputRef := range pipeline.OutputRefs {
			if output, found := outputs[outputRef]; found {
				for _, inputRefs := range pipeline.InputRefs {
					output.AddInputFrom(inputs[inputRefs])
				}
			}

		}
	}
	return pipeline
}

func addPostFilters(p *Pipeline) {

	var postFilters []string
	if viaq.HasJournalSource(p.inputSpecs) {
		postFilters = append(postFilters, viaq.ViaqJournal)
		p.filterMap[viaq.ViaqJournal] = filter.InternalFilterSpec{
			FilterSpec:        &obs.FilterSpec{Type: viaq.ViaqJournal},
			SuppliesTransform: true,
			TranformFactory: func(id string, inputs ...string) framework.Element {
				return viaq.NewJournal(id, inputs...)
			},
		}
	}

	postFilters = append(postFilters, viaq.Viaq)
	p.filterMap[viaq.Viaq] = filter.InternalFilterSpec{
		FilterSpec:        &obs.FilterSpec{Type: viaq.Viaq},
		SuppliesTransform: true,
		TranformFactory: func(id string, inputs ...string) framework.Element {
			// Build all log_source VRL
			return viaq.New(id, inputs, p.inputSpecs)
		},
	}
	if viaq.HasContainerSource(p.inputSpecs) {
		postFilters = append(postFilters, viaq.ViaqDedot)
		p.filterMap[viaq.ViaqDedot] = filter.InternalFilterSpec{
			FilterSpec:        &obs.FilterSpec{Type: viaq.ViaqDedot},
			SuppliesTransform: true,
			TranformFactory: func(id string, inputs ...string) framework.Element {
				return viaq.NewDedot(id, inputs...)
			},
		}
	}
	p.FilterRefs = append(p.FilterRefs, postFilters...)
}

func (p *Pipeline) Name() string {
	if p.PipelineSpec.Name == "" {
		return helpers.MakeID("pipeline", strconv.Itoa(p.index))
	}
	return p.PipelineSpec.Name
}

func (p *Pipeline) initFilter(index int, filterRef string) {
	names := sets.NewString()
	if f, ok := p.filterMap[filterRef]; ok {
		filterID := helpers.MakeID(filterRef, strconv.Itoa(index))
		if pf := NewPipelineFilter(p.Name(), filterID, f, p.PipelineSpec); pf != nil {
			names.Insert(pf.ID())
			if len(p.Filters) > 0 {
				last := p.Filters[len(p.Filters)-1]
				pf.AddInputFrom(last)
			}
			p.Filters = append(p.Filters, pf)
		}
	}
}

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
