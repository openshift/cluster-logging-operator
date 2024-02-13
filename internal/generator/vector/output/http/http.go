package http

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize/schema/otel"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultHttpTimeoutSecs = 10
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

type HttpBatch struct {
	ComponentID string
	MaxBytes    string
}

func (b HttpBatch) Name() string {
	return "vectorHttpBatch"
}

func (b HttpBatch) Template() string {
	return `{{define "` + b.Name() + `" -}}
[sinks.{{.ComponentID}}.batch]
max_bytes = {{.MaxBytes}}
{{end}}`
}

type HttpRequest struct {
	ComponentID string
	Timeout     string
	Headers     Element
}

func (h HttpRequest) Name() string {
	return "vectorHttpRequest"
}

func (h HttpRequest) Template() string {
	return `{{define "` + h.Name() + `" -}}
[sinks.{{.ComponentID}}.request]
timeout_secs = {{.Timeout}}
{{kv .Headers -}}
{{end}}`
}

func Normalize(id string, inputs []string) Element {
	removeFile := `del(.file)`
	return Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         removeFile,
	}
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	normalizeID := vectorhelpers.MakeID(id, "normalize")
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Normalize(normalizeID, inputs),
			Debug(helpers.MakeID(id, "debug"), normalizeID),
		}
	}
	var els []Element
	if op.Has(constants.AnnotationEnableSchema) && o.Http != nil && o.Http.Schema == constants.OTELSchema {
		schemaID := vectorhelpers.MakeID(id, "otel")
		els = append(els, otel.Transform(schemaID, inputs))
		inputs = []string{schemaID}
	}
	els = append(els, Normalize(normalizeID, inputs))
	sink := Output(id, o, []string{dedottedID}, secret, op)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(

		els,
		[]Element{
			normalize.DedotLabels(dedottedID, []string{normalizeID}),
			sink,
			Encoding(id),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
			Request(id, o, strategy),
		},
		common.TLS(id, o, secret, op),
		BasicAuth(id, o, secret),
		BearerTokenAuth(id, o, secret),
	)
}

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) *Http {
	return &Http{
		ComponentID: id,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.URL,
		Method:      Method(o.Http),
		RootMixin:   common.NewRootMixin(nil),
	}
}

func Method(h *logging.Http) string {
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

func Request(id string, o logging.OutputSpec, strategy common.ConfigStrategy) *common.Request {
	timeout := DefaultHttpTimeoutSecs
	if o.Http != nil && o.Http.Timeout != 0 {
		timeout = o.Http.Timeout
	}
	req := common.NewRequest(id, strategy)
	req.TimeoutSecs.Value = timeout
	if o.Http != nil && len(o.Http.Headers) != 0 {
		req.SetHeaders(o.Http.Headers)
	}
	return req
}

func Encoding(id string) Element {
	return HttpEncoding{
		ComponentID: id,
		Codec:       httpEncodingJson,
	}
}

func BasicAuth(id string, o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: id,
		})
		if common.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				Username: common.GetFromSecret(secret, constants.ClientUsername),
				Password: common.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}

func BearerTokenAuth(id string, o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if secret != nil {
		// Inject token from secret, either provided by user using a custom secret
		// or from the default logcollector service account.
		if common.HasBearerTokenFileKey(secret) {
			conf = append(conf, BasicAuthConf{
				Desc:        "Bearer Auth Config",
				ComponentID: id,
			}, BearerToken{
				Token: common.GetFromSecret(secret, constants.BearerTokenFileKey),
			})
		}
	}
	return conf
}
