package elasticsearch

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type Elasticsearch struct {
	IDKey       genhelper.OptionalPair
	Desc        string
	ComponentID string
	Inputs      string
	Index       string
	Endpoint    string
	Version     int
	common.RootMixin
}

func (e Elasticsearch) Name() string {
	return "elasticsearchTemplate"
}

func (e Elasticsearch) Template() string {
	return `{{define "` + e.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "elasticsearch"
inputs = {{.Inputs}}
endpoints = ["{{.Endpoint}}"]
{{.IDKey}}
bulk.index = "{{ .Index }}"
bulk.action = "create"
{{.Compression}}
{{- if ne .Version 0 }}
api_version = "v{{ .Version }}"
{{- end }}
{{end}}`
}

func (e *Elasticsearch) SetCompression(algo string) {
	e.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets helpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, helpers.MakeInputs(inputs...)),
		}
	}
	outputs := []Element{}
	if o.Elasticsearch.Version == 6 {
		addID := helpers.MakeID(id, "add_id")
		outputs = append(outputs, Remap{
			ComponentID: addID,
			Inputs:      helpers.MakeInputs(inputs...),
			VRL: `._id = encode_base64(uuid_v4())
if exists(.kubernetes.event.metadata.uid) {
  ._id = .kubernetes.event.metadata.uid
}`,
		})
		inputs = []string{addID}
	}
	sink := Output(id, o, inputs, secrets, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}

	outputs = append(outputs,
		sink,
		common.NewEncoding(id, ""),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op, Option{Name: URL, Value: o.Elasticsearch.URL}),
		auth.HTTPAuth(id, o.Elasticsearch.Authentication, secrets),
	)

	return outputs
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets helpers.Secrets, op Options) *Elasticsearch {
	idKey := genhelper.NewOptionalPair("id_key", nil)
	if o.Elasticsearch.Version == 6 {
		idKey.Value = "_id"
	}
	es := Elasticsearch{
		ComponentID: id,
		IDKey:       idKey,
		Endpoint:    o.Elasticsearch.URL,
		Inputs:      helpers.MakeInputs(inputs...),
		Index:       o.Elasticsearch.Index,
		RootMixin:   common.NewRootMixin(nil),
		Version:     o.Elasticsearch.Version,
	}
	return &es
}
