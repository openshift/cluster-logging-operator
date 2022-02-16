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
	clfb                          *ClusterLogForwarderBuilder
	inputName                     string
	enableMultilineErrorDetection bool
	input                         *logging.InputSpec
	pipelineName                  string
}

type InputSpecVisitor func(spec *logging.InputSpec)
type OutputSpecVisiter func(spec *logging.OutputSpec)
type PipelineSpecVisitor func(spec *logging.PipelineSpec)

func NewClusterLogForwarderBuilder(clf *logging.ClusterLogForwarder) *ClusterLogForwarderBuilder {
	return &ClusterLogForwarderBuilder{
		Forwarder: clf,
	}
}

func (b *ClusterLogForwarderBuilder) FromInput(inputName string) *PipelineBuilder {
	pipelineBuilder := &PipelineBuilder{
		clfb:      b,
		inputName: inputName,
		input:     &logging.InputSpec{Name: inputName},
	}
	return pipelineBuilder
}
func (b *ClusterLogForwarderBuilder) FromInputWithVisitor(inputName string, visit InputSpecVisitor) *PipelineBuilder {
	pipelineBuilder := b.FromInput(inputName)
	visit(pipelineBuilder.input)
	return pipelineBuilder
}

func (p *PipelineBuilder) WithMultineErrorDetection() *PipelineBuilder {
	p.enableMultilineErrorDetection = true
	return p
}

func (p *PipelineBuilder) Named(name string) *PipelineBuilder {
	p.pipelineName = name
	return p
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

func (p *PipelineBuilder) ToKafkaOutput() *ClusterLogForwarderBuilder {
	return p.ToOutputWithVisitor(func(output *logging.OutputSpec) {}, logging.OutputTypeKafka)
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
				URL:  "http://0.0.0.0:9200",
			}
		case logging.OutputTypeSyslog:
			output = &logging.OutputSpec{
				Name: logging.OutputTypeSyslog,
				Type: logging.OutputTypeSyslog,
				URL:  "tcp://0.0.0.0:24224",
				OutputTypeSpec: logging.OutputTypeSpec{
					Syslog: &logging.Syslog{},
				},
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
		case logging.OutputTypeKafka:
			output = &logging.OutputSpec{
				Name: logging.OutputTypeKafka,
				Type: logging.OutputTypeKafka,
				URL:  "https://0.0.0.0:9093",
				Secret: &logging.OutputSecretSpec{
					Name: "kafka",
				},
			}
		default:
			output = &logging.OutputSpec{
				Name: outputName,
			}
		}
		visit(output)

		clf.Spec.Outputs = append(clf.Spec.Outputs, *output)
	}

	if p.input != nil {
		clf.Spec.Inputs = append(clf.Spec.Inputs, *p.input)
	}

	added := false
	pipelineName := forwardPipelineName
	if p.pipelineName != "" {
		pipelineName = p.pipelineName
	}
	clf.Spec.Pipelines, added = addInputOutputToPipeline(p.inputName, output.Name, pipelineName, clf.Spec.Pipelines)
	if !added {
		clf.Spec.Pipelines = append(clf.Spec.Pipelines, logging.PipelineSpec{
			Name:                  pipelineName,
			InputRefs:             []string{p.inputName},
			OutputRefs:            []string{output.Name},
			DetectMultilineErrors: p.enableMultilineErrorDetection,
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
