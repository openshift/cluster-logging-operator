package fluentd

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace, name string, resNames *factory.ForwarderResourceNames, op framework.Options) []framework.Section {
	return []framework.Section{
		{
			Header(op),
			`Generated fluentd conf Header`,
		},
		{
			Sources(clspec, clfspec, namespace, op),
			"Set of all input sources",
		},
		{
			Concat(clfspec, op),
			`Concat pipeline section`,
		},
		{
			Ingress(clfspec, op),
			"Ingress pipeline",
		},
		// input ends
		// give a hook here
		{
			InputsToPipeline(clfspec, op),
			"Inputs go to pipelines",
		},
		{
			PipelineToOutputs(clfspec, op),
			"Pipeline to Outputs",
		},
		// output begins here
		// give a hook here
		{
			Outputs(clspec, secrets, clfspec, op),
			"Outputs",
		},
	}
}

func Header(op framework.Options) []framework.Element {
	const Header = `
{{define "header" -}}
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>
{{end}}
`
	return []framework.Element{
		framework.ConfLiteral{
			TemplateName: "header",
			TemplateStr:  Header,
		},
	}
}
