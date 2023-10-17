package splunk

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
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
	Index        string
}

func (s Splunk) Name() string {
	return "SplunkVectorTemplate"
}

func (s Splunk) Template() string {
	return `{{define "` + s.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "splunk_hec_logs"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
compression = "none"
default_token = "{{.DefaultToken}}"
timestamp_key = "@timestamp"
{{if .Index -}}
index = "{{.Index}}"
{{end -}} 
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
	id := vectorhelpers.FormatComponentID(o.Name)
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	dedottedID := normalize.ID(id, "dedot")
	return MergeElements(
		[]Element{
			normalize.DedotLabels(dedottedID, inputs),
			Output(o, []string{dedottedID}, secret, op),
			Encoding(o),
			output.NewBuffer(id),
			output.NewRequest(id),
		},
		TLSConf(o, secret, op),
	)
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {
	splunk := Splunk{
		ComponentID:  vectorhelpers.FormatComponentID(o.Name),
		Inputs:       vectorhelpers.MakeInputs(inputs...),
		Endpoint:     o.URL,
		DefaultToken: security.GetFromSecret(secret, constants.SplunkHECTokenKey),
	}
	if o.Splunk != nil && strings.TrimSpace(o.Splunk.Index) != "" {
		splunk.Index = o.Splunk.Index
	}
	return splunk
}

func Encoding(o logging.OutputSpec) Element {
	return SplunkEncoding{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Codec:       splunkEncodingJson,
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret, op Options) []Element {
	if tlsConf := security.GenerateTLSConf(o, secret, op, false); tlsConf != nil {
		tlsConf.NeedsEnabled = false
		return []Element{tlsConf}
	}
	return []Element{}
}
