package functional

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

const (
	forwardOutputName   = "fluentForward"
	forwardPipelineName = "forward-pipeline"
)

type ClusterLogForwarderBuilder struct {
	Forwarder *logging.ClusterLogForwarder
}

type PipelineBuilder struct {
	clfb      *ClusterLogForwarderBuilder
	inputName string
}

type OutputSpecVisiter func(spec *logging.OutputSpec)

func NewClusterLogForwarderBuilder(clf *logging.ClusterLogForwarder) *ClusterLogForwarderBuilder {
	return &ClusterLogForwarderBuilder{
		Forwarder: clf,
	}
}

func (b *ClusterLogForwarderBuilder) FromInput(inputName string) *PipelineBuilder {
	pipelineBuilder := &PipelineBuilder{
		clfb:      b,
		inputName: inputName,
	}
	return pipelineBuilder
}

func (p *PipelineBuilder) ToFluentForwardOutput() *ClusterLogForwarderBuilder {
	return p.ToFluentForwardOutputWithVisitor(func(output *logging.OutputSpec) {})
}

func (p *PipelineBuilder) ToFluentForwardOutputWithVisitor(visit OutputSpecVisiter) *ClusterLogForwarderBuilder {
	clf := p.clfb.Forwarder
	outputs := clf.Spec.OutputMap()
	var output *logging.OutputSpec
	var found bool
	if output, found = outputs[forwardOutputName]; !found {
		output = &logging.OutputSpec{
			Name: forwardOutputName,
			Type: logging.OutputTypeFluentdForward,
			URL:  "tcp://0.0.0.0:24224",
		}
		visit(output)
		clf.Spec.Outputs = append(clf.Spec.Outputs, *output)
	}
	added := false
	clf.Spec.Pipelines, added = addInputToPipeline(p.inputName, forwardPipelineName, clf.Spec.Pipelines)
	if !added {
		clf.Spec.Pipelines = append(clf.Spec.Pipelines, logging.PipelineSpec{
			Name:       forwardPipelineName,
			InputRefs:  []string{p.inputName},
			OutputRefs: []string{output.Name},
		})
	}
	return p.clfb
}

func addInputToPipeline(inputName, pipelineName string, pipelineSpecs []logging.PipelineSpec) ([]logging.PipelineSpec, bool) {
	pipelines := []logging.PipelineSpec{}
	found := false
	for _, pipeline := range pipelineSpecs {
		if pipelineName == pipeline.Name {
			found = true
			pipeline.InputRefs = append(pipeline.InputRefs, inputName)
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, found
}
