package splunk

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
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

func (s *Splunk) SetCompression(algo string) {
	s.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	timestampID := vectorhelpers.MakeID(id, "timestamp")

	var indexTemplate Element
	splunkSink := sink(id, o, []string{timestampID}, "", secrets, op)
	if hasIndexKey(o.Splunk) {
		splunkIndexID := vectorhelpers.MakeID(id, "splunk_index")
		indexTemplate = commontemplate.TemplateRemap(splunkIndexID, []string{timestampID}, o.Splunk.Index, splunkIndexID, "Splunk Index")
		splunkSink = sink(id, o, []string{splunkIndexID}, splunkIndexID, secrets, op)
	}

	if strategy != nil {
		strategy.VisitSink(splunkSink)
	}
	return []Element{
		FixTimestampFormat(timestampID, inputs),
		indexTemplate,
		splunkSink,
		common.NewEncoding(id, common.CodecJSON),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, index string, secrets observability.Secrets, op Options) *Splunk {
	s := &Splunk{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Endpoint:    o.Splunk.URL,
		Index:       Tenant(o.Splunk, index),
		RootMixin:   common.NewRootMixin("none"),
	}
	authentication := o.Splunk.Authentication
	if authentication != nil && authentication.Token != nil {
		s.DefaultToken = vectorhelpers.SecretFrom(authentication.Token)
	}
	return s
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

func hasIndexKey(s *obs.Splunk) bool {
	return s != nil && s.Index != ""
}

func Tenant(s *obs.Splunk, index string) Element {
	if !hasIndexKey(s) {
		return Nil
	}
	return KV("index", fmt.Sprintf(`"{{ ._internal.%s }}"`, index))
}
