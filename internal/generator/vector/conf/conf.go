package conf

import (
	"sort"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/metrics"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/pipeline"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	corev1 "k8s.io/api/core/v1"
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
func Conf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace, forwarderName string, resNames *factory.ForwarderResourceNames, op framework.Options) []framework.Section {

	// Init inputs, outputs, pipelines
	inputMap := map[string]*input.Input{}
	inputCompMap := map[string]helpers.InputComponent{}
	for _, i := range clfspec.Inputs {
		a := input.NewInput(i, namespace, resNames, op)
		inputMap[i.Name] = a
		inputCompMap[i.Name] = a
	}

	outputMap := map[string]*output.Output{}
	for _, spec := range clfspec.Outputs {
		o := output.NewOutput(spec, secrets, op)
		outputMap[spec.Name] = o
	}

	filters := filter.NewInternalFilterMap(clfspec.FilterMap())
	pipelineMap := map[string]*pipeline.Pipeline{}
	for i, p := range clfspec.Pipelines {
		a := pipeline.NewPipeline(i, p, inputCompMap, outputMap, filters, clfspec.Inputs)
		pipelineMap[p.Name] = a
	}

	// generate sections, deferring input wiring to config generation
	sections := framework.Section{}
	for _, i := range sortAdapters(inputMap) {
		sections.Elements = append(sections.Elements, i.Elements()...)
	}
	for _, p := range sortAdapters(pipelineMap) {
		sections.Elements = append(sections.Elements, p.Elements()...)
	}
	for _, o := range sortAdapters(outputMap) {
		sections.Elements = append(sections.Elements, o.Elements()...)
	}

	minTlsVersion, cipherSuites := op.TLSProfileInfo(logging.OutputSpec{}, ",")
	return []framework.Section{
		{
			Global(namespace, forwarderName),
			`vector global options`,
		},
		{
			Elements: source.MetricsSources(source.InternalMetricsSourceName),
		},
		sections,
		{
			Elements: []framework.Element{
				metrics.AddNodeNameToMetric(metrics.AddNodenameToMetricTransformName, []string{source.InternalMetricsSourceName}),
				metrics.PrometheusOutput(metrics.PrometheusOutputSinkName, []string{metrics.AddNodenameToMetricTransformName}, minTlsVersion, cipherSuites),
			},
		},
	}

}

// sortAdapters sorts ClusterLogForwarder adapters to ensure consistent generation of component configs
func sortAdapters[V *input.Input | *pipeline.Pipeline | *output.Output](m map[string]V) []V {
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
