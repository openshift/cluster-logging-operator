package otelc

import (
	"sort"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/processors"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/output"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

func NewConfig(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op utils.Options) (config *api.Config) {

	config = api.NewConfig()

	inputsToReceivers := map[string]api.Receivers{}
	hasContainerInputs := false
	for _, i := range clfspec.Inputs {
		a := adapters.NewInput(i)
		if a.IsContainerSource() {
			hasContainerInputs = true
		}
		receivers := input.New(a, resNames, secrets, op)
		inputsToReceivers[i.Name] = receivers
		config.AddReceivers(receivers)
	}

	outputsToExporters := map[string]api.Exporters{}
	for _, spec := range clfspec.Outputs {
		o := adapters.NewOutput(spec)
		// TODO fix inputs to the exporters
		exporters, extensions := output.New(o, clfspec.Inputs, secrets, op)
		outputsToExporters[spec.Name] = exporters
		config.AddExporters(exporters)
		config.AddExtensions(extensions)
	}

	pipelineProcessors := []string{}
	// enrich container records with the pod labels from the Kubernetes API;
	// the CRI-O input operators already set the association attributes from
	// the log file path
	if hasContainerInputs {
		k8sattrs := newK8sAttributesProcessor()
		config.Processors.Add(k8sattrs.ID(), k8sattrs)
		pipelineProcessors = append(pipelineProcessors, k8sattrs.ID())
	}
	// openshift labels are static values from the labels filter, carried as
	// resource attributes by the resource processor
	if resource := newOpenshiftLabelsProcessor(clfspec.Filters); resource != nil {
		config.Processors.Add(resource.ID(), resource)
		pipelineProcessors = append(pipelineProcessors, resource.ID())
	}
	sort.Strings(pipelineProcessors)

	for _, p := range clfspec.Pipelines {

		pipeline := api.NewLogsPipeline(p.Name)
		for _, inputRefs := range p.InputRefs {
			receivers := inputsToReceivers[inputRefs]
			for _, rec := range receivers {
				pipeline.Receivers = append(pipeline.Receivers, rec.ID())
			}
		}
		for _, outputRefs := range p.OutputRefs {
			exporters := outputsToExporters[outputRefs]
			for _, ex := range exporters {
				pipeline.Exporters = append(pipeline.Exporters, ex.ID())
			}
		}
		pipeline.Processors = append(pipeline.Processors, pipelineProcessors...)
		sort.Strings(pipeline.Receivers)
		sort.Strings(pipeline.Exporters)
		config.Service.AddPipeline(pipeline)
	}

	return config
}

// newK8sAttributesProcessor creates the k8sattributes processor that enriches
// container records with pod labels as k8s.pod.label.<key> resource
// attributes. Records are associated to their pod via the resource attributes
// the CRI-O input operators extract from the log file path.
func newK8sAttributesProcessor() *processors.K8sAttributes {
	k8sattrs := processors.NewK8sAttributes("")
	k8sattrs.Extract = &processors.ExtractConfig{
		Metadata: []string{
			"k8s.namespace.name",
			"k8s.pod.name",
			"k8s.pod.uid",
		},
		Labels: []processors.FieldExtract{
			{
				// every pod label, named after the data model:
				// k8s.pod.label.<label key>. The tag name is a regex back
				// reference expanded by the processor at runtime; the dollar
				// sign is doubled because the collector's config loader
				// otherwise treats $... as an environment variable reference
				TagName:  "k8s.pod.label.$$1",
				KeyRegex: "(.*)",
				From:     "pod",
			},
		},
	}
	k8sattrs.PodAssociation = []processors.Association{
		{
			Sources: []processors.AssociationSource{{From: "resource_attribute", Name: "k8s.pod.uid"}},
		},
		{
			Sources: []processors.AssociationSource{
				{From: "resource_attribute", Name: "k8s.pod.name"},
				{From: "resource_attribute", Name: "k8s.namespace.name"},
			},
		},
	}
	return k8sattrs
}

// newOpenshiftLabelsProcessor creates the resource processor that carries the
// static openshift labels of a labels filter as openshift.label.<key> resource
// attributes, or nil when no labels filter is defined
func newOpenshiftLabelsProcessor(filterSpecs []obs.FilterSpec) *processors.Resource {
	for _, f := range filterSpecs {
		if f.Type != obs.FilterTypeOpenshiftLabels {
			continue
		}
		resource := processors.NewResource("openshift_labels")
		keys := make([]string, 0, len(f.OpenshiftLabels))
		for key := range f.OpenshiftLabels {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			resource.SetAttribute("openshift.label."+key, f.OpenshiftLabels[key])
		}
		if len(resource.Attributes) == 0 {
			return nil
		}
		return resource
	}
	return nil
}
