package http

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize/schema/otel"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultHttpMaxBytes    = 100 * 1024 // 100 KB
	DefaultHttpTimeoutSecs = 10

	NormalizeHttp = "normalize_http"
)

var (
	httpEncodingJson = fmt.Sprintf("%q", "json")
)

type Http struct {
	ComponentID string
	Inputs      string
	URI         string
	Method      string
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
{{end}}
`
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

func Normalize(componentID string, inputs []string) Element {
	removeFile := `del(.file)`
	return Remap{
		ComponentID: componentID,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         removeFile,
	}
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	id := helpers.FormatComponentID(o.Name)
	component := strings.ToLower(vectorhelpers.Replacer.Replace(fmt.Sprintf("%s_%s", o.Name, NormalizeHttp)))
	dedottedID := normalize.ID(id, "dedot")
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Normalize(component, inputs),
			Debug(strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)), component),
		}
	}
	return MergeElements(
		Schema(o, id, component, inputs, op),
		[]Element{
			normalize.DedotLabels(dedottedID, []string{component}),
			Output(o, []string{dedottedID}, secret, op),
			Encoding(o),
			output.NewBuffer(id),
			Request(o),
		},
		security.TLS(o, secret, op),
		BasicAuth(o, secret),
		BearerTokenAuth(o, secret),
	)
}

func Schema(o logging.OutputSpec, outputName, component string, inputs []string, op Options) []Element {
	if op.Has(constants.AnnotationEnableSchema) && o.Http != nil && o.Http.Schema == constants.OTELSchema {
		schemaID := otel.ID(outputName, "otel")
		return []Element{
			otel.Transform(schemaID, inputs),
			Normalize(component, []string{schemaID}),
		}
	}
	return []Element{
		Normalize(component, inputs),
	}
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {
	return Http{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Inputs:      vectorhelpers.MakeInputs(inputs...),
		URI:         o.URL,
		Method:      Method(o.Http),
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

func Request(o logging.OutputSpec) *output.Request {
	timeout := DefaultHttpTimeoutSecs
	if o.Http != nil && o.Http.Timeout != 0 {
		timeout = o.Http.Timeout
	}
	req := output.NewRequest(strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)))
	req.TimeoutSecs.Value = timeout
	if o.Http != nil && len(o.Http.Headers) != 0 {
		req.SetHeaders(o.Http.Headers)
	}
	return req
}

func Encoding(o logging.OutputSpec) Element {
	return HttpEncoding{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Codec:       httpEncodingJson,
	}
}

func BasicAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		})
		if security.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				Username: security.GetFromSecret(secret, constants.ClientUsername),
				Password: security.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}

func BearerTokenAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if secret != nil {
		// Inject token from secret, either provided by user using a custom secret
		// or from the default logcollector service account.
		if security.HasBearerTokenFileKey(secret) {
			conf = append(conf, BasicAuthConf{
				Desc:        "Bearer Auth Config",
				ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
			}, BearerToken{
				Token: security.GetFromSecret(secret, constants.BearerTokenFileKey),
			})
		}
	}
	return conf
}
