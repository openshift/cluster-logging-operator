package observability

import (
	"net/url"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	forwardPipelineName = "forward-pipeline"
)

type ClusterLogForwarderBuilder struct {
	Forwarder   *obs.ClusterLogForwarder
	inputSpecs  map[string]*obs.InputSpec
	filterSpecs map[string]*obs.FilterSpec
}

type PipelineBuilder struct {
	clfb         *ClusterLogForwarderBuilder
	inputName    string
	input        *obs.InputSpec
	pipelineName string
	filterName   string
	filter       *obs.FilterSpec
	inputRefs    []string
}

type InputSpecVisitor func(spec *obs.InputSpec)
type OutputSpecVisitor func(spec *obs.OutputSpec)
type FilterSpecVisitor func(spec *obs.FilterSpec)
type PipelineSpecVisitor func(spec *obs.PipelineSpec)

func NewClusterLogForwarderBuilder(clf *obs.ClusterLogForwarder) *ClusterLogForwarderBuilder {
	return &ClusterLogForwarderBuilder{
		Forwarder:   clf,
		inputSpecs:  map[string]*obs.InputSpec{},
		filterSpecs: map[string]*obs.FilterSpec{},
	}
}

func (b *ClusterLogForwarderBuilder) FromInput(inputType obs.InputType, visitors ...InputSpecVisitor) *PipelineBuilder {
	visitors = append([]InputSpecVisitor{func(spec *obs.InputSpec) {
		spec.Type = inputType
		switch inputType {
		case obs.InputTypeApplication:
			spec.Application = &obs.Application{}
		case obs.InputTypeInfrastructure:
			spec.Infrastructure = &obs.Infrastructure{
				Sources: obs.InfrastructureSources,
			}
		case obs.InputTypeAudit:
			spec.Audit = &obs.Audit{
				Sources: obs.AuditSources,
			}
		}
	}}, visitors...)
	return b.FromInputName(string(inputType), visitors...)
}

func (b *ClusterLogForwarderBuilder) FromInputName(inputName string, visitors ...InputSpecVisitor) *PipelineBuilder {
	inputSpec := &obs.InputSpec{Name: inputName}
	for _, v := range visitors {
		v(inputSpec)
	}
	if _, ok := b.inputSpecs[inputName]; ok {
		inputSpec = b.inputSpecs[inputName]
	}
	pipelineBuilder := &PipelineBuilder{
		clfb:      b,
		inputName: inputSpec.Name,
		input:     inputSpec,
		inputRefs: []string{},
	}
	b.inputSpecs[inputName] = inputSpec
	return pipelineBuilder
}

func (p *PipelineBuilder) AndInput(inputType obs.InputType) *PipelineBuilder {
	return p.AndInputName(string(inputType))
}

func (p *PipelineBuilder) AndInputName(inputName string) *PipelineBuilder {
	p.inputRefs = append(p.inputRefs, inputName)
	return p
}

func (p *PipelineBuilder) WithFilter(filterName string, visitors ...FilterSpecVisitor) *PipelineBuilder {
	builder := p.clfb
	filterSpec := &obs.FilterSpec{Name: filterName}
	if _, ok := builder.filterSpecs[filterName]; ok {
		filterSpec = builder.filterSpecs[filterName]
	}
	p.filterName = filterName
	p.filter = filterSpec

	for _, v := range visitors {
		v(p.filter)
	}

	builder.filterSpecs[filterName] = filterSpec
	return p
}

func (p *PipelineBuilder) WithMultilineErrorDetectionFilter() *PipelineBuilder {
	p.WithFilter(string(obs.FilterTypeDetectMultiline), func(spec *obs.FilterSpec) {
		spec.Type = obs.FilterTypeDetectMultiline
	})
	return p
}

func (p *PipelineBuilder) WithParseJson() *PipelineBuilder {
	p.WithFilter(string(obs.FilterTypeParse), func(spec *obs.FilterSpec) {
		spec.Type = obs.FilterTypeParse
	})
	return p
}

// Named is the name to be given to the ClusterLogForwarder pipeline
func (p *PipelineBuilder) Named(name string) *PipelineBuilder {
	p.pipelineName = name
	return p
}

func (p *PipelineBuilder) ToElasticSearchOutput(visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeElasticsearch)
		output.Type = obs.OutputTypeElasticsearch
		output.Elasticsearch = &obs.Elasticsearch{
			URLSpec: obs.URLSpec{
				URL: "http://0.0.0.0:9200",
			},
			Index:   `{.log_type||"notfound"}-write`,
			Version: 8,
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeElasticsearch))
}

func (p *PipelineBuilder) ToSyslogOutput(rfc obs.SyslogRFCType, visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeSyslog)
		output.Type = obs.OutputTypeSyslog
		output.Syslog = &obs.Syslog{
			RFC: rfc,
			URL: "tcp://127.0.0.1:24224",
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeSyslog))
}

func (p *PipelineBuilder) ToCloudwatchOutput(auth obs.CloudwatchAuthentication, visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeCloudwatch)
		output.Type = obs.OutputTypeCloudwatch
		output.TLS = &obs.OutputTLSSpec{
			InsecureSkipVerify: true,
		}
		output.Cloudwatch = &obs.Cloudwatch{
			URL:            "https://localhost:5000",
			Region:         "us-east-1",
			GroupName:      `group-prefix.{.log_type||"none"}`,
			Authentication: &auth,
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeCloudwatch))
}

func (p *PipelineBuilder) ToLokiOutput(lokiURL url.URL, visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeLoki)
		output.Type = obs.OutputTypeLoki
		output.Loki = &obs.Loki{
			URLSpec: obs.URLSpec{
				URL: lokiURL.String(),
			},
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeLoki))
}

func (p *PipelineBuilder) ToSplunkOutput(hecTokenSecret obs.SecretReference, visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeSplunk)
		output.Type = obs.OutputTypeSplunk
		output.Splunk = &obs.Splunk{
			URLSpec: obs.URLSpec{
				URL: "http://localhost:8088",
			},
			Authentication: &obs.SplunkAuthentication{
				Token: &hecTokenSecret,
			},
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeSplunk))
}

func (p *PipelineBuilder) ToKafkaOutput(visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	kafkaVisitor := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeKafka)
		output.Type = obs.OutputTypeKafka
		output.Kafka = &obs.Kafka{
			URL:   "https://localhost:9093",
			Topic: kafka.AppLogsTopic,
		}
		output.TLS = &obs.OutputTLSSpec{
			TLSSpec: obs.TLSSpec{
				Key: &obs.SecretReference{
					Key:        constants.ClientPrivateKey,
					SecretName: kafka.DeploymentName,
				},
				Certificate: &obs.ValueReference{
					Key:        constants.ClientCertKey,
					SecretName: kafka.DeploymentName,
				},
				CA: &obs.ValueReference{
					Key:        constants.TrustedCABundleKey,
					SecretName: kafka.DeploymentName,
				},
			},
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(kafkaVisitor, string(obs.OutputTypeKafka))
}

func (p *PipelineBuilder) ToHttpOutput(visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeHTTP)
		output.Type = obs.OutputTypeHTTP
		output.HTTP = &obs.HTTP{
			URLSpec: obs.URLSpec{
				URL: "http://localhost:8090",
			},
			Method: "POST",
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeHTTP))
}

func (p *PipelineBuilder) ToOtlpOutput(visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeOTLP)
		output.Type = obs.OutputTypeOTLP
		output.OTLP = &obs.OTLP{
			URL: "http://localhost:4318/v1/logs",
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeOTLP))
}
func (p *PipelineBuilder) ToAzureMonitorOutput(visitors ...func(output *obs.OutputSpec)) *ClusterLogForwarderBuilder {
	v := func(output *obs.OutputSpec) {
		output.Name = string(obs.OutputTypeAzureMonitor)
		output.Type = obs.OutputTypeAzureMonitor
		output.TLS = &obs.OutputTLSSpec{
			InsecureSkipVerify: true,
		}
		output.AzureMonitor = &obs.AzureMonitor{
			LogType: "myLogType",
			Host:    "acme.com:3000",
			Authentication: &obs.AzureMonitorAuthentication{
				SharedKey: &obs.SecretReference{
					Key:        constants.SharedKey,
					SecretName: "azure-secret",
				},
			},
		}
		for _, v := range visitors {
			v(output)
		}
	}
	return p.ToOutputWithVisitor(v, string(obs.OutputTypeAzureMonitor))
}

func (p *PipelineBuilder) ToOutputWithVisitor(visit OutputSpecVisitor, outputName string) *ClusterLogForwarderBuilder {
	clf := p.clfb.Forwarder
	outputs := internalobs.Outputs(clf.Spec.Outputs).Map()
	var output obs.OutputSpec
	var found bool
	if output, found = outputs[outputName]; !found {
		switch outputName {
		default:
			output = obs.OutputSpec{
				Name: outputName,
			}
		}
		visit(&output)

		clf.Spec.Outputs = append(clf.Spec.Outputs, output)
	}

	if p.input != nil {
		found = false
		for _, input := range clf.Spec.Inputs {
			if input.Name == p.input.Name {
				found = true
			}
		}
		if !found {
			clf.Spec.Inputs = append(clf.Spec.Inputs, *p.input)
		}
	}

	if p.filter != nil {
		found = false
		for _, filter := range clf.Spec.Filters {
			if filter.Name == p.filter.Name {
				found = true
			}
		}
		if !found {
			clf.Spec.Filters = append(clf.Spec.Filters, *p.filter)
		}
	}

	added := false
	pipelineName := forwardPipelineName
	if p.pipelineName != "" {
		pipelineName = p.pipelineName
	}
	clf.Spec.Pipelines, added = addInputOutputToPipeline(p.inputName, output.Name, pipelineName, clf.Spec.Pipelines)
	if !added {
		inputRefs := sets.NewString(p.inputRefs...)
		inputRefs.Insert(p.inputName)
		pSpec := obs.PipelineSpec{
			Name:       pipelineName,
			InputRefs:  inputRefs.List(),
			OutputRefs: []string{output.Name},
		}
		if p.filterName != "" {
			pSpec.FilterRefs = append(pSpec.FilterRefs, p.filterName)
		}
		clf.Spec.Pipelines = append(clf.Spec.Pipelines, pSpec)
	}
	return p.clfb
}

func addInputOutputToPipeline(inputName, outputName, pipelineName string, pipelineSpecs []obs.PipelineSpec) ([]obs.PipelineSpec, bool) {
	pipelines := []obs.PipelineSpec{}
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
