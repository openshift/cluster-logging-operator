package pipeline

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	openshiftfilter "github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"os"
	"strconv"
)

// Pipeline is an adapter between logging API and config generation
type Pipeline struct {
	logging.PipelineSpec
	index     int
	filterMap map[string]filter.InternalFilterSpec
	Filters   []*PipelineFilter
}

func (o *Pipeline) Elements() []framework.Element {
	elements := []framework.Element{}
	for _, pf := range o.Filters {
		elements = append(elements, pf.Element())
	}
	return elements
}

func NewPipeline(index int, p logging.PipelineSpec, inputs map[string]*input.Input, outputs map[string]*output.Output, filters map[string]*filter.InternalFilterSpec) *Pipeline {
	pipeline := &Pipeline{
		PipelineSpec: p,
		index:        index,
		filterMap:    map[string]filter.InternalFilterSpec{},
	}
	for name, f := range filters {
		pipeline.filterMap[name] = *f
	}

	addPrefilters(pipeline)

	for i, filterName := range pipeline.FilterRefs {
		pipeline.initFilter(i, filterName)
	}
	if len(pipeline.FilterRefs) > 0 {
		if len(pipeline.Filters) == 0 {
			log.V(0).Info("Runtime error in pipelineAdapter while processing filters.  Filters spec'd but not constructed", "filterRefs", pipeline.FilterRefs)
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
			output := outputs[outputRef]
			for _, inputRefs := range pipeline.InputRefs {
				output.AddInputFrom(inputs[inputRefs])
			}
		}
	}
	return pipeline
}

// TODO: add migration to treat like any other
func addPrefilters(p *Pipeline) {
	prefilters := []string{}
	if p.DetectMultilineErrors {
		p.filterMap[openshiftfilter.DetectMultilineException] = filter.InternalFilterSpec{
			FilterSpec:        &logging.FilterSpec{Type: openshiftfilter.DetectMultilineException},
			SuppliesTransform: true,
			TranformFactory:   openshiftfilter.NewDetectException,
		}
		prefilters = append(prefilters, openshiftfilter.DetectMultilineException)
	}
	if len(p.Labels) > 0 {
		p.filterMap[openshiftfilter.Labels] = filter.InternalFilterSpec{
			FilterSpec: &logging.FilterSpec{Type: openshiftfilter.Labels}, Labels: p.Labels}
		prefilters = append(prefilters, openshiftfilter.Labels)
	}
	if p.Parse == openshiftfilter.ParseTypeJson {
		p.filterMap[openshiftfilter.ParseJson] = filter.InternalFilterSpec{
			FilterSpec: &logging.FilterSpec{Type: openshiftfilter.ParseJson}}
		prefilters = append(prefilters, openshiftfilter.ParseJson)
	}
	p.FilterRefs = append(prefilters, p.FilterRefs...)
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

// PipelineFilter is an adapter between CLF pipeline filter instance and config generation
type PipelineFilter struct {
	ids  []string
	Next []helpers.Component
	vrl  string

	//transformFactory is a function that takes inputs and returns a transform
	transformFactory func(string) framework.Element
}

func (pf *PipelineFilter) ID() string {
	return pf.ids[0]
}
func (pf *PipelineFilter) InputIDs() []string {
	return pf.ids
}

func (pf *PipelineFilter) AddInputFrom(n helpers.Component) {
	pf.Next = append(pf.Next, n)
}

func NewPipelineFilter(pipelineName, filterRef string, spec filter.InternalFilterSpec) *PipelineFilter {
	ids := []string{helpers.MakePipelineID(pipelineName, filterRef)}
	if spec.SuppliesTransform {
		return &PipelineFilter{
			ids: ids,
			transformFactory: func(inputs string) framework.Element {
				return spec.TranformFactory(ids[0], inputs)
			},
		}
	}
	if vrl, err := filter.VRLFrom(&spec); err != nil {
		log.Error(err, "bad filter", "filterRef", filterRef, "spec.type", spec.Type, "spec.Name", spec.Name)
		return nil
	} else {
		return &PipelineFilter{
			ids: ids,
			vrl: vrl,
		}
	}
}
func (o *PipelineFilter) Element() framework.Element {
	inputs := []string{}
	for _, n := range o.Next {
		inputs = append(inputs, n.InputIDs()...)
	}
	if o.transformFactory != nil {
		return o.transformFactory(helpers.MakeInputs(inputs...))
	}
	return elements.Remap{
		ComponentID: o.ids[0],
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         o.vrl,
	}
}
