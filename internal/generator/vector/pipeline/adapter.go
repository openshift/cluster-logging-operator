package pipeline

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"os"
	"strconv"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
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

func NewPipeline(index int, p obs.PipelineSpec, inputs map[string]helpers.InputComponent, outputs map[string]*output.Output, filters map[string]*filter.InternalFilterSpec, inputSpecs []obs.InputSpec, addPostFilters func(p *Pipeline)) *Pipeline {
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

func AddPostFilters(p *Pipeline) {

	postFilters := []string{v1.Viaq}
	p.filterMap[v1.Viaq] = filter.InternalFilterSpec{
		FilterSpec:        &obs.FilterSpec{Type: v1.Viaq},
		SuppliesTransform: true,
		TranformFactory: func(id string, inputs ...string) framework.Element {
			// Build all log_source VRL
			return v1.New(id, inputs, p.inputSpecs)
		},
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
