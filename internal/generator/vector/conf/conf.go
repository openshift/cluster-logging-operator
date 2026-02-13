package conf

import (
	"sort"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/metrics"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/pipeline"
	corev1 "k8s.io/api/core/v1"
)

const (
	InternalMetricsSourceName = "internal_metrics"
)

// Design of next generation conf generation:
// * 1:1 Wrappers/representation (e.g. Node) of CLF types
//   Inputs are  1:* outflow
//   Outputs are *:1 inflow
//   Pipeline input* -> (filter->filter)* ->*output
// * Actual config representing a CLF type is 1+ elements
//   - "internal" id and "routing" for elements of a given CLF type are crafted from the name, unique, more or less arbitrary
//   - Input/Pipeline final element is ID crafted from well-formed name used for "inputs" to other components
// * ID:
// * input_<name>(_<element_purpose>)*    (e.g. input_application, input_application_dedot)
// * output_<name>(_<element_purpose>)*    (e.g. output_mykafka, output_mykafka_dedot)
// * pipeline_<name>(_<element_purpose>)*

/*
   spec:
     filters:
     - name: foo
       type: kubeAPIAudit
       kubeAPIAudit:
         omitResponseCodes: [201]
     outputs:
     - name: mykafka:
       type: kafka
     pipelinelines:
     - name: apipe
       outputRefs: ["mykafka"]
       inputRefs: ["application"]
       filterRefs: ["foo"]

    [source.input_application]
    [transforms.input_application_container]
      input_application
    [transforms.pipeline_apipe_foo_kubeapiaudit]
       input_application_container
	[sinks.output_mykafka_dedot]
       pipeline_apipe_foo_kubeapiaudit
   	[sinks.output_mykafka]
       output_mykafka_dedot
*/
//nolint:govet // using declarative style
func Conf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) []framework.Section {
	op[helpers.CLFSpec] = internalobs.ClusterLogForwarderSpec(clfspec)

	// Init inputs, outputs, pipelines
	inputMap := map[string]*internalobs.Input{}
	inputCompMap := map[string]helpers.InputComponent{}
	for _, i := range clfspec.Inputs {
		a := internalobs.NewInput(i)
		inputMap[i.Name] = a
		inputCompMap[i.Name] = a
	}

	outputMap := map[string]*adapters.Output{}
	op[framework.OptionForwarderName] = forwarderName
	for _, spec := range clfspec.Outputs {
		o := adapters.NewOutput(spec)
		outputMap[spec.Name] = o
	}

	filters := filter.NewInternalFilterMap(internalobs.FilterMap(clfspec))
	pipelineMap := map[string]*pipeline.Pipeline{}
	for i, p := range clfspec.Pipelines {
		a := pipeline.NewPipeline(i, p, inputCompMap, outputMap, filters, clfspec.Inputs, pipeline.AddSystemFilters)
		pipelineMap[p.Name] = a
	}

	// generate sections, deferring input wiring to config generation
	sections := framework.Section{}
	for _, i := range sortAdapters(inputMap) {
		elements, ids := input.NewSource(i, resNames, secrets, op)
		i.Ids = ids
		sections.Elements = append(sections.Elements, elements...)
	}
	for _, p := range sortAdapters(pipelineMap) {
		sections.Elements = append(sections.Elements, p.Elements()...)
	}
	for _, o := range sortAdapters(outputMap) {
		sections.Elements = append(sections.Elements, output.New(o, o.InputIDs, secrets, o, op)...)
	}

	minTlsVersion, cipherSuites := framework.TLSProfileInfo(op, obs.OutputSpec{}, ",")
	return []framework.Section{
		{
			Global(namespace, forwarderName),
			`vector global options`,
		},
		{
			Elements: []framework.Element{
				api.NewConfig(func(c *api.Config) {
					c.Sources[InternalMetricsSourceName] = sources.NewInternalMetrics()
				}),
			},
		},
		sections,
		{
			Elements: []framework.Element{
				metrics.AddNodeNameToMetric(metrics.AddNodenameToMetricTransformName, []string{InternalMetricsSourceName}),
				metrics.PrometheusOutput(metrics.PrometheusOutputSinkName, []string{metrics.AddNodenameToMetricTransformName}, minTlsVersion, cipherSuites, op),
			},
		},
	}

}

// sortAdapters sorts ClusterLogForwarder adapters to ensure consistent generation of component configs
func sortAdapters[V *internalobs.Input | *pipeline.Pipeline | *adapters.Output](m map[string]V) []V {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := []V{}
	for _, k := range keys {
		result = append(result, m[k])
	}
	return result
}
