package adapters

import (
	"os"
	"strconv"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	v1 "github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// Pipeline is an adapter between logging API and config generation
type Pipeline struct {
	obs.PipelineSpec
	index      int
	filterMap  map[string]InternalFilterSpec
	Filters    []*PipelineFilter
	inputSpecs []obs.InputSpec
}

// Transforms creates instances of transforms based upon the pipeline spec
func (p *Pipeline) Transforms() (tfs api.Transforms) {
	tfs = api.Transforms{}
	for _, pf := range p.Filters {
		tfs.Add(pf.ID(), pf.Transform())
	}
	return tfs
}

func NewPipeline(index int, p obs.PipelineSpec, inputs map[string]helpers.InputComponent, outputs map[string]*Output, filters map[string]*InternalFilterSpec, inputSpecs []obs.InputSpec, addPostFilters func(p *Pipeline)) *Pipeline {
	pipeline := &Pipeline{
		PipelineSpec: p,
		index:        index,
		filterMap:    map[string]InternalFilterSpec{},
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
			os.Exit(1)
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

func AddSystemFilters(p *Pipeline) {
	postFilterName := v1.Viaq
	p.filterMap[postFilterName] = InternalFilterSpec{
		FilterSpec: &obs.FilterSpec{Type: v1.Viaq},
		Factory: func(inputs ...string) types.Transform {
			return v1.New(inputs, p.inputSpecs)
		},
	}

	var preFilters []string
	var otherFilters []string

	for _, refName := range p.FilterRefs {
		spec, exists := p.filterMap[refName]
		isEarlyStage := exists && (spec.Type == obs.FilterTypeParse || spec.Type == obs.FilterTypeDetectMultiline)
		if isEarlyStage {
			preFilters = append(preFilters, refName)
		} else {
			otherFilters = append(otherFilters, refName)
		}
	}

	// [Parse/Multiline] -> [Viaq] -> [Others]
	newRefs := make([]string, 0, len(preFilters)+1+len(otherFilters))

	newRefs = append(newRefs, preFilters...)
	newRefs = append(newRefs, postFilterName)
	newRefs = append(newRefs, otherFilters...)

	p.FilterRefs = newRefs
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
		if pf := NewPipelineFilter(p.Name(), filterID, f); pf != nil {
			names.Insert(pf.ID())
			if len(p.Filters) > 0 {
				last := p.Filters[len(p.Filters)-1]
				pf.AddInputFrom(last)
			}
			p.Filters = append(p.Filters, pf)
		}
	}
}
