package common

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
)

type Request struct {
	ComponentID   string
	RetryAttempts int
	Concurrency   helpers.OptionalPair
	TimeoutSecs   helpers.OptionalPair
	headers       map[string]string
}

// NewRequest section for an output
// Ref: LOG-4536 for RetryAttempts default
func NewRequest(id string) *Request {
	return &Request{
		ComponentID:   id,
		RetryAttempts: 17,
		Concurrency:   helpers.NewOptionalPair("concurrency", nil),
		TimeoutSecs:   helpers.NewOptionalPair("timeout_secs", nil),
	}
}

func (r *Request) Name() string {
	return "request"
}

func (r *Request) Template() string {
	return `{{define "` + r.Name() + `" -}}
[sinks.{{.ComponentID}}.request]
retry_attempts = {{.RetryAttempts}}
{{ .Concurrency -}}
{{ .TimeoutSecs }}
{{kv .Headers }}
{{end}}
`
}

func (r *Request) Headers() elements.KeyVal {
	return elements.KV("headers", utils.ToHeaderStr(r.headers, "%q=%q"))
}

func (r *Request) SetHeaders(headers map[string]string) {
	r.headers = headers
}
