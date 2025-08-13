package http

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
)

type Http struct {
	ComponentID  string
	Inputs       string
	URI          string
	Method       string
	Proxy        string
	LinePerEvent bool
	common.RootMixin
}

func (h Http) Name() string {
	return "vectorHttpTemplate"
}

func (h Http) Template() string {
	return `{{define "` + h.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "http"
inputs = {{.Inputs}}
uri = "{{.URI}}"
method = "{{.Method}}"
{{with .LinePerEvent}}
framing.method = "newline_delimited"
{{end}}
{{with .Proxy -}}
proxy.enabled = true
proxy.http = "{{.}}"
proxy.https = "{{.}}"
{{end -}}
{{.Compression}}
{{end}}
`
}

func (h *Http) SetCompression(algo string) {
	h.Compression.Value = algo
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(helpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	var els []Element
	sink := Output(id, o, inputs, secrets, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(

		els,
		[]Element{
			sink,
			common.NewEncoding(id, common.CodecJSON),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			Request(id, o, strategy),
			tls.New(id, o.TLS, secrets, op),
			auth.HTTPAuth(id, o.HTTP.Authentication, secrets, op),
		},
	)
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, op Options) *Http {
	return &Http{
		ComponentID:  id,
		Inputs:       vectorhelpers.MakeInputs(inputs...),
		URI:          o.HTTP.URL,
		Method:       Method(o.HTTP),
		Proxy:        o.HTTP.ProxyURL,
		LinePerEvent: o.HTTP.LinePerEvent,
		RootMixin:    common.NewRootMixin(nil),
	}
}

func Method(h *obs.HTTP) string {
	if h == nil {
		return "post"
	}
	switch h.Method {
	case "GET":
		return "get"
	case "HEAD":
		return "head"
	case "POST":
		return "post"
	case "PUT":
		return "put"
	case "DELETE":
		return "delete"
	case "OPTIONS":
		return "options"
	case "TRACE":
		return "trace"
	case "PATCH":
		return "patch"
	}
	return "post"
}

func Request(id string, o obs.OutputSpec, strategy common.ConfigStrategy) *common.Request {
	req := common.NewRequest(id, strategy)
	if o.HTTP != nil && o.HTTP.Timeout != 0 {
		req.TimeoutSecs.Value = o.HTTP.Timeout
	}
	if o.HTTP != nil && len(o.HTTP.Headers) != 0 {
		req.SetHeaders(o.HTTP.Headers)
	}
	return req
}
