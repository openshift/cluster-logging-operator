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
	ID_Key      string
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
bulk.index = "{{ .Index }}"
bulk.action = "create"
id_key = "_id"
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
	sink := Output(id, o, inputs, secrets, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	outputs := []Element{
		sink,
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		tls.New(id, o.TLS, secrets, op),
		auth.HTTPAuth(id, o.Elasticsearch.Authentication, secrets),
	}

	return outputs
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets helpers.Secrets, op Options) *Elasticsearch {
	es := Elasticsearch{
		ComponentID: id,
		Endpoint:    o.Elasticsearch.URL,
		Inputs:      helpers.MakeInputs(inputs...),
		Index:       o.Elasticsearch.Index,
		RootMixin:   common.NewRootMixin(nil),
		Version:     o.Elasticsearch.Version,
	}
	return &es
}
