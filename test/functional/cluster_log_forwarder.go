package functional

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
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
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, logging.OutputTypeFluentdForward)
}

func (p *PipelineBuilder) ToElasticSearchOutput() *ClusterLogForwarderBuilder {
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, logging.OutputTypeElasticsearch)
}

func (p *PipelineBuilder) ToSyslogOutput() *ClusterLogForwarderBuilder {
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, logging.OutputTypeSyslog)
}

func (p *PipelineBuilder) ToCloudwatchOutput() *ClusterLogForwarderBuilder {
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, logging.OutputTypeCloudwatch)
}

func (p *PipelineBuilder) ToOutputWithVisitor(visit OutputSpecVisiter, outputName string) *ClusterLogForwarderBuilder {
	clf := p.clfb.Forwarder
	outputs := clf.Spec.OutputMap()
	var output *logging.OutputSpec
	var found bool
	if output, found = outputs[outputName]; !found {
		switch outputName {
		case logging.OutputTypeFluentdForward:
			output = &logging.OutputSpec{
				Name: logging.OutputTypeFluentdForward,
				Type: logging.OutputTypeFluentdForward,
				URL:  "tcp://0.0.0.0:24224",
			}
		case logging.OutputTypeElasticsearch:
			output = &logging.OutputSpec{
				Name: logging.OutputTypeElasticsearch,
				Type: logging.OutputTypeElasticsearch,
				URL:  "https://0.0.0.0:9200",
			}
		case logging.OutputTypeSyslog:
			output = &logging.OutputSpec{
				Name: logging.OutputTypeSyslog,
				Type: logging.OutputTypeSyslog,
				URL:  "tcp://0.0.0.0:24224",
			}
		case logging.OutputTypeCloudwatch:
			groupPrefix := "group-prefix"
			output = &logging.OutputSpec{
				Name: logging.OutputTypeCloudwatch,
				Type: logging.OutputTypeCloudwatch,
				URL:  "https://localhost:5000",
				OutputTypeSpec: logging.OutputTypeSpec{
					Cloudwatch: &logging.Cloudwatch{
						Region:      "us-east-1",
						GroupBy:     logging.LogGroupByLogType,
						GroupPrefix: &groupPrefix,
					},
				},
				Secret: &logging.OutputSecretSpec{
					Name: "cloudwatch",
				},
			}
		}

		visit(output)
		clf.Spec.Outputs = append(clf.Spec.Outputs, *output)
	}
	added := false
	clf.Spec.Pipelines, added = addInputOutputToPipeline(p.inputName, output.Name, forwardPipelineName, clf.Spec.Pipelines)
	if !added {
		clf.Spec.Pipelines = append(clf.Spec.Pipelines, logging.PipelineSpec{
			Name:       forwardPipelineName,
			InputRefs:  []string{p.inputName},
			OutputRefs: []string{output.Name},
		})
	}
	return p.clfb
}

func addInputOutputToPipeline(inputName, outputName, pipelineName string, pipelineSpecs []logging.PipelineSpec) ([]logging.PipelineSpec, bool) {
	pipelines := []logging.PipelineSpec{}
	found := false
	for _, pipeline := range pipelineSpecs {
		if pipelineName == pipeline.Name {
			found = true

			outputRefs := sets.NewString(pipeline.OutputRefs...)
			if !outputRefs.Has(outputName) {
				pipeline.OutputRefs = append(pipeline.OutputRefs, outputName)
			}
			inputRefs := sets.NewString(pipeline.InputRefs...)
			if !inputRefs.Has(inputName) {
				pipeline.InputRefs = append(pipeline.InputRefs, inputName)
			}
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, found
}
