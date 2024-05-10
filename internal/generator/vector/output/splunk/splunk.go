package splunk

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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
index = "{{.Index}}"
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

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	timestampID := vectorhelpers.MakeID(id, "timestamp")

	splunkSink := sink(id, o, []string{timestampID}, secrets, op)
	if strategy != nil {
		strategy.VisitSink(splunkSink)
	}
	return []Element{
		FixTimestampFormat(timestampID, inputs),
		splunkSink,
		Encoding(id, o),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
	}
}

func sink(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options) *Splunk {
	s := &Splunk{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		Endpoint:    o.Splunk.URL,
		Index:       o.Splunk.Index,
		RootMixin:   common.NewRootMixin("none"),
	}
	if o.Splunk.Authentication != nil {
		s.DefaultToken = secrets.AsString(o.Splunk.Authentication.Token)
	}
	return s
}

func Encoding(id string, o obs.OutputSpec) Element {
	return SplunkEncoding{
		ComponentID: id,
		Codec:       splunkEncodingJson,
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
