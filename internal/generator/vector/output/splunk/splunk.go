package splunk

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
)

var (
	splunkEncodingJson = fmt.Sprintf("%q", "json")
)

type Splunk struct {
	ComponentID  string
	Inputs       string
	Endpoint     string
	DefaultToken string
}

func (s Splunk) Name() string {
	return "SplunkVectorTemplate"
}

func (s Splunk) Template() string {
	return `{{define "` + s.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "splunk_hec"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
compression = "none"
default_token = "{{.DefaultToken}}"
{{end}}`
}

type SplunkEncoding struct {
	ComponentID string
	Codec       string
}

func (se SplunkEncoding) Name() string {
	return "splunkEncoding"
}

func (se SplunkEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = {{.Codec}}
{{end}}`
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	outputName := vectorhelpers.FormatComponentID(o.Name)
	dedottedID := normalize.ID(outputName, "dedot")
	return MergeElements(
		[]Element{
			normalize.DedotLabels(dedottedID, inputs),
			Output(o, []string{dedottedID}, secret, op),
			Encoding(o),
		},
		TLSConf(o, secret, op),
	)
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {
	return Splunk{
		ComponentID:  strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Inputs:       vectorhelpers.MakeInputs(inputs...),
		Endpoint:     o.URL,
		DefaultToken: security.GetFromSecret(secret, "hecToken"),
	}
}

func Encoding(o logging.OutputSpec) Element {
	return SplunkEncoding{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Codec:       splunkEncodingJson,
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret, op Options) []Element {
	if tlsConf := security.GenerateTLSConf(o, secret, op, false); tlsConf != nil {
		return []Element{tlsConf}
	}

	return []Element{}
}
