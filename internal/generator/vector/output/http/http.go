package http

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/auth"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
)

var (
	httpEncodingJson = fmt.Sprintf("%q", "json")
)

type Http struct {
	ComponentID string
	Inputs      string
	URI         string
	Method      string
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
{{.Compression}}
{{end}}
`
}

func (h *Http) SetCompression(algo string) {
	h.Compression.Value = algo
}

type HttpEncoding struct {
	ComponentID string
	Codec       string
}

func (h HttpEncoding) Name() string {
	return "vectorHttpEncoding"
}

func (h HttpEncoding) Template() string {
	return `{{define "` + h.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = {{.Codec}}
{{end}}`
}

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
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
			Encoding(id),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			Request(id, o, strategy),
			tls.New(id, o.TLS, secrets, op),
			auth.HTTPAuth(id, o.HTTP.Authentication, secrets),
		},
	)
}

func Output(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, op Options) *Http {
	return &Http{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.HTTP.URL,
		Method:      Method(o.HTTP),
		RootMixin:   common.NewRootMixin(nil),
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

func Encoding(id string) Element {
	return HttpEncoding{
		ComponentID: id,
		Codec:       httpEncodingJson,
	}
}
