package splunk

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
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
	Index        Element
	common.RootMixin
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
{{.Compression}}
default_token = "{{.DefaultToken}}"
{{kv .Index -}}
timestamp_key = "@timestamp"
{{end}}`
}

type SplunkEncoding struct {
	ComponentID  string
	Codec        string
	ExceptFields Element
}

func (se SplunkEncoding) Name() string {
	return "splunkEncoding"
}

func (se SplunkEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = {{.Codec}}
{{kv .ExceptFields -}}
{{end}}`
}

func (s *Splunk) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	componentID := vectorhelpers.MakeID(id, "add_splunk_index")
	timestampID := vectorhelpers.MakeID(id, "timestamp")
	dedottedID := vectorhelpers.MakeID(id, "dedot")

	dedotInputs := inputs
	indexRemapElement := SetSplunkIndexRemap(o.Splunk, componentID, inputs)
	if len(indexRemapElement) != 0 {
		dedotInputs = []string{componentID}
	}
	sink := Output(id, o, []string{timestampID}, secret, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(
		indexRemapElement,
		[]Element{
			normalize.DedotLabels(dedottedID, dedotInputs),
			FixTimestampFormat(timestampID, []string{dedottedID}),
			sink,
			Encoding(id, o),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			common.NewRequest(id, strategy),
		},
		common.TLS(id, o, secret, op),
	)
}

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) *Splunk {
	return &Splunk{
		ComponentID:  id,
		Inputs:       vectorhelpers.MakeInputs(inputs...),
		Endpoint:     o.URL,
		DefaultToken: common.GetFromSecret(secret, constants.SplunkHECTokenKey),
		Index:        AddSplunkIndexToSink(o.Splunk),
		RootMixin:    common.NewRootMixin("none"),
	}
}

func hasCustomIndex(s *logging.Splunk) bool {
	return s != nil && (s.IndexKey != "" || s.IndexName != "")
}

func SetSplunkIndexRemap(s *logging.Splunk, componentID string, inputs []string) []Element {
	var vrl string
	var index string

	if !hasCustomIndex(s) {
		return []Element{}
	}

	switch {
	// If key is not found, a write index of "" writes to default index defined in Splunk
	case s.IndexKey != "":
		vrl = `
val = .%s
if !is_null(val) {
	.write_index = val
} else {
	.write_index = ""
}
`
		index = s.IndexKey

	case s.IndexName != "":
		vrl = `
.write_index = "%s"
`
		index = s.IndexName
	}
	return []Element{
		Remap{
			Desc:        "Set Splunk Index",
			ComponentID: componentID,
			Inputs:      vectorhelpers.MakeInputs(inputs...),
			VRL:         strings.TrimSpace(fmt.Sprintf(vrl, index)),
		},
	}
}

func AddSplunkIndexToSink(s *logging.Splunk) Element {
	if !hasCustomIndex(s) {
		return Nil
	}

	return KV("index", fmt.Sprintf("%q", "{{ write_index }}"))
}

func AddSplunkEncodeExceptFields(s *logging.Splunk) Element {
	if !hasCustomIndex(s) {
		return Nil
	}

	return KV("except_fields", "[\"write_index\", \"_internal\"]")
}

func Encoding(id string, o logging.OutputSpec) Element {
	return SplunkEncoding{
		ComponentID:  id,
		Codec:        splunkEncodingJson,
		ExceptFields: AddSplunkEncodeExceptFields(o.Splunk),
	}
}

func FixTimestampFormat(componentID string, inputs []string) Element {
	var vrl = `
ts, err = parse_timestamp(.@timestamp,"%+")
if err != null {
	log("could not parse timestamp. err=" + err, rate_limit_secs: 0)
} else {
	.@timestamp = ts
}
`
	return Remap{
		Desc:        "Ensure timestamp field well formatted for Splunk",
		ComponentID: componentID,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		VRL:         vrl,
	}
}
