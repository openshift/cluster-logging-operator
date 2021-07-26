package fluentd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrNoValidInputs    = errors.New("No valid inputs found in ClusterLogForwarder")
	ErrNoOutputs        = errors.New("No outputs defined in ClusterLogForwarder")
	ErrInvalidOutputURL = func(o logging.OutputSpec) error {
		return fmt.Errorf("Invalid URL in %s output in ClusterLogForwarder", o.Name)
	}
	ErrInvalidInput = errors.New("Invalid Input")
)

//nolint:govet // using declarative style
func Conf(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op Options) []Section {
	return []Section{
		{
			Header(op),
			`Generated fluentd conf Header`,
		},
		{
			Sources(clfspec, op),
			"Set of all input sources",
		},
		{
			PrometheusMetrics(clfspec, op),
			"Section to add measurement, and dispatch to Concat or Ingress pipelines",
		},
		{
			Concat(clfspec, op),
			`Concat pipeline 
			section`,
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

func Header(op Options) []Element {
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
	return []Element{
		ConfLiteral{
			TemplateName: "header",
			TemplateStr:  Header,
		},
	}
}

func Verify(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op Options) error {
	var err error
	types := GatherSources(clfspec, op)
	types = AddLegacySources(types, op)
	if !types.HasAny(logging.InputNameApplication, logging.InputNameInfrastructure, logging.InputNameAudit) {
		return ErrNoValidInputs
	}
	if len(clfspec.Outputs) == 0 &&
		!IsIncludeLegacyForwardConfig(op) &&
		!IsIncludeLegacySyslogConfig(op) {
		return ErrNoOutputs
	}
	for _, p := range clfspec.Pipelines {
		if _, err := json.Marshal(p.Labels); err != nil {
			return ErrInvalidInput
		}
	}
	for _, o := range clfspec.Outputs {
		if _, err := url.Parse(o.URL); err != nil {
			return ErrInvalidOutputURL(o)
		}
	}
	return err
}
