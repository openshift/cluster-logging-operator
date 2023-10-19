package next

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/outputs"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/source"
)

var _ = Describe("Vector Config Generation", func() {
	var (
		spec = logging.ClusterLogForwarderSpec{
			Inputs: []logging.InputSpec{
				{Name: logging.InputNameInfrastructure},
				{Name: logging.InputNameAudit},
			},
			Outputs: []logging.OutputSpec{
				{Name: logging.OutputTypeFluentdForward, Type: logging.OutputTypeFluentdForward, URL: "url: tcp://0.0.0.0:24224"},
				{Name: logging.OutputTypeElasticsearch, Type: logging.OutputTypeElasticsearch, URL: "http:/0.0.0.0:9200"},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "forward-pipeline",
					InputRefs:  []string{logging.InputNameInfrastructure, logging.InputNameAudit},
					OutputRefs: []string{logging.OutputTypeElasticsearch, logging.OutputTypeFluentdForward},
					//FilterRefs: []string{"viaq"},
				},
			},
		}
	)
	FIt("", func() {
		//inputMap := MapInputsToComponentsFrom(spec)
		for _, c := range MapMe(spec) {
			fmt.Printf("%s\n", c)
			for _, i := range c.Inputs() {
				fmt.Printf("|-->%s\n", i)
			}
		}
		Expect("")
		Expect(true).To(BeFalse())
	})
})

//
//func MapInputsToComponentsFrom(spec logging.ClusterLogForwarderSpec) map[string]*sets.String {
//	outputsMap := spec.OutputMap()
//	//filterMap := spec.FilterMap()
//	results := map[string]*sets.String{}
//	sourceTypes := generator.GatherSources(&spec, nil)
//	for _, pipeline := range spec.Pipelines {
//		inputs := []string{}
//		//all input refs
//		for _, inputRef := range pipeline.InputRefs {
//			inputs = append(inputs, source.MakeIDsFor(inputRef, sourceTypes)...)
//		}
//		for _, filterRef := range pipeline.FilterRefs {
//			//registry.NewFilter(inputRefs
//			filter := registry.LookupProto(lastRef, filterMap)
//		}
//		if len(pipeline.FilterRefs) > 0 {
//			//lastRef := pipeline.FilterRefs[len(pipeline.FilterRefs)-1]
//			//filter := registry.LookupProto(lastRef, filterMap)
//			//name = filter_<pipeline_name>_filter[_special?]
//		} else {
//
//		}
//		for _, ref := range pipeline.OutputRefs {
//			output := outputsMap[ref]
//			id := outputs.MakeID(*output)
//			paths := results[id]
//			if paths == nil {
//				paths = &sets.String{}
//				results[id] = paths
//			}
//			paths.Insert(inputs...)
//		}
//	}
//	return results
//}

var (
	inputSpecApplication    = &logging.InputSpec{Name: logging.InputNameApplication}
	inputSpecContainer      = &logging.InputSpec{Name: logging.InputNameContainer}
	inputSpecInfrastructure = &logging.InputSpec{Name: logging.InputNameInfrastructure}
	inputSpecNode           = &logging.InputSpec{Name: logging.InputNameNode}
	inputSpecAudit          = &logging.InputSpec{Name: logging.InputNameAudit}
)

func PipelineInputMap(pipeline logging.PipelineSpec, spec logging.ClusterLogForwarderSpec) map[*logging.PipelineSpec][]*logging.InputSpec {
	inputs := spec.InputMap()
	results := map[*logging.PipelineSpec][]*logging.InputSpec{}
	for _, pipeline := range spec.Pipelines {
		inputSpecs := []*logging.InputSpec{}
		for _, ref := range pipeline.InputRefs {
			if input, found := inputs[ref]; found {
				inputSpecs = append(inputSpecs, input)
			} else {
				switch ref {
				case logging.InputNameApplication:
					inputSpecs = append(inputSpecs, inputSpecApplication)
				case logging.InputNameContainer:
					inputSpecs = append(inputSpecs, inputSpecContainer)
				case logging.InputNameInfrastructure:
					inputSpecs = append(inputSpecs, inputSpecInfrastructure)
				case logging.InputNameNode:
					inputSpecs = append(inputSpecs, inputSpecNode)
				case logging.InputNameAudit:
					inputSpecs = append(inputSpecs, inputSpecAudit)
				}
			}
		}
		results[&pipeline] = inputSpecs
	}
	return results
}

func MapMe(spec logging.ClusterLogForwarderSpec) ComponentSet {
	components := ComponentSet{}
	inputs := map[string]Component{}
	outputs := map[string]Component{}
	filterSpecs := spec.FilterMap()
	filterSpecs["viaq"] = &logging.FilterSpec{}
	for _, input := range spec.Inputs {
		source := NewSourceConfig(input)
		components.Add(source)
		inputs[source.Name()] = source
	}
	for _, output := range spec.Outputs {
		sink := NewSinkConfig(output)
		components.Add(sink)
		outputs[sink.Name()] = sink
	}
	for _, pipeline := range spec.Pipelines {
		for _, outRef := range pipeline.OutputRefs {
			if len(pipeline.FilterRefs) > 0 {

			} else {

			}
			for _, inRef := range pipeline.InputRefs {
				input := inputs[inRef]
				outputs[outRef].AddInput(input)
			}
		}

	}
	return components
}

type Component interface {
	ID() string
	Name() string
	AddInput(Component)
	Inputs() ComponentSet
}

type ComponentSet map[Component]Component

func (s ComponentSet) Add(comp Component) Component {
	if c, found := s[comp]; !found {
		s[comp] = comp
		return comp
	} else {
		return c
	}

}

type SourceConfig struct {
	input logging.InputSpec
}

func NewSourceConfig(i logging.InputSpec) *SourceConfig {
	return &SourceConfig{
		input: i,
	}
}

func (s *SourceConfig) AddInput(Component) {
}
func (s *SourceConfig) Inputs() ComponentSet {
	return ComponentSet{}
}

func (s *SourceConfig) String() string {
	return s.ID()
}

func (s *SourceConfig) ID() string {
	return source.MakeID(s.input.Name)
}
func (s *SourceConfig) Name() string {
	return s.input.Name
}

type TransformConfig struct {
	pipeline   logging.PipelineSpec
	filterName string
	filterSpec logging.FilterSpec
	inputs     ComponentSet
}

func NewTransformConfig(pipeline logging.PipelineSpec, filterName string, filterSpec logging.FilterSpec) *TransformConfig {
	return &TransformConfig{
		pipeline:   pipeline,
		filterName: filterName,
		filterSpec: filterSpec,
		inputs:     ComponentSet{},
	}
}

type SinkConfig struct {
	output logging.OutputSpec
	inputs ComponentSet
}

func NewSinkConfig(o logging.OutputSpec) *SinkConfig {
	return &SinkConfig{
		output: o,
		inputs: ComponentSet{},
	}
}

func (s *SinkConfig) AddInput(comp Component) {
	s.inputs[comp] = comp
}
func (s *SinkConfig) ID() string {
	return outputs.MakeID(s.output)
}
func (s *SinkConfig) Name() string {
	return s.output.Name
}
func (s *SinkConfig) String() string {
	return s.ID()
}
func (s *SinkConfig) Inputs() ComponentSet {
	return s.inputs
}
