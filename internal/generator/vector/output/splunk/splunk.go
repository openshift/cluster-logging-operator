package splunk

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
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
type = "splunk_hec_logs"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
compression = "none"
default_token = "{{.DefaultToken}}"
timestamp_key = "@timestamp"
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
	return New(id, o, inputs, secret, op)
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	return MergeElements(
		[]Element{
			normalize.DedotLabels(dedottedID, inputs),
			Output(id, o, []string{dedottedID}, secret, op),
			Encoding(id, o),
			common.NewBuffer(id),
			common.NewRequest(id),
		},
		TLSConf(id, o, secret, op),
	)
}

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {
	return Splunk{
		ComponentID:  id,
		Inputs:       vectorhelpers.MakeInputs(inputs...),
		Endpoint:     o.URL,
		DefaultToken: common.GetFromSecret(secret, constants.SplunkHECTokenKey),
	}
}

func Encoding(id string, o logging.OutputSpec) Element {
	return SplunkEncoding{
		ComponentID: id,
		Codec:       splunkEncodingJson,
	}
}

func TLSConf(id string, o logging.OutputSpec, secret *corev1.Secret, op Options) []Element {
	if tlsConf := common.GenerateTLSConfWithID(id, o, secret, op, false); tlsConf != nil {
		tlsConf.NeedsEnabled = false
		return []Element{tlsConf}
	}
	return []Element{}
}
