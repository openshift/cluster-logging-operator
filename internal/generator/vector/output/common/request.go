package common

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
)

type Request struct {
	ComponentID            string
	RetryAttempts          helpers.OptionalPair
	RetryInitialBackoffSec helpers.OptionalPair
	RetryMaxDurationSec    helpers.OptionalPair
	Concurrency            helpers.OptionalPair
	TimeoutSecs            helpers.OptionalPair
	headers                map[string]string
}

// NewRequest section for an output
// Ref: LOG-4536 for RetryAttempts default
func NewRequest(id string, s ConfigStrategy) *Request {
	r := Request{
		ComponentID:            id,
		RetryAttempts:          helpers.NewOptionalPair("retry_attempts", nil),
		RetryInitialBackoffSec: helpers.NewOptionalPair("retry_initial_backoff_secs", nil),
		RetryMaxDurationSec:    helpers.NewOptionalPair("retry_max_duration_secs", nil),
		Concurrency:            helpers.NewOptionalPair("concurrency", nil),
		TimeoutSecs:            helpers.NewOptionalPair("timeout_secs", nil),
	}
	if s != nil {
		r = s.VisitRequest(r)
	}
	return &r
}

func (r *Request) Name() string {
	return "request"
}

func (r *Request) isEmpty() bool {
	return len(r.headers) == 0 && r.RetryInitialBackoffSec.String()+
		r.RetryMaxDurationSec.String()+
		r.RetryAttempts.String()+
		r.Concurrency.String()+
		r.TimeoutSecs.String() == ""
}

func (r *Request) Template() string {
	if r.isEmpty() {
		return `{{define "` + r.Name() + `" -}}{{end}}`
	}
	return `{{define "` + r.Name() + `" -}}
[sinks.{{.ComponentID}}.request]
{{ .RetryAttempts }}
{{ .RetryInitialBackoffSec }}
{{ .RetryMaxDurationSec }}
{{ .Concurrency }}
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
