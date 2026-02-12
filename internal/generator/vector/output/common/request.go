package common

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
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

func NewApiRequest(o observability.TunableOutput) (r *sinks.Request) {
	var duration time.Duration
	t := o.GetTuning()
	if t.MinRetryDuration != nil && t.MinRetryDuration.Seconds() > 0 {
		r = &sinks.Request{}
		// time.Duration is default nanosecond. Convert to seconds first.
		duration = *t.MinRetryDuration * time.Second
		r.RetryInitialBackoffSecs = uint(duration.Seconds())
	}
	if t.MaxRetryDuration != nil && t.MaxRetryDuration.Seconds() > 0 {
		if r == nil {
			r = &sinks.Request{}
		}
		duration = *t.MaxRetryDuration * time.Second
		r.RetryMaxDurationSec = uint(duration.Seconds())
	}
	return r
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
	return elements.KV("headers", toHeaderStr(r.headers, "%q=%q"))
}

func (r *Request) SetHeaders(headers map[string]string) {
	r.headers = headers
}

func toHeaderStr(h map[string]string, formatStr string) string {
	if len(h) == 0 {
		return ""
	}
	sortedKeys := make([]string, len(h))
	i := 0
	for k := range h {
		sortedKeys[i] = k
		i += 1
	}
	sort.Strings(sortedKeys)
	hv := make([]string, len(h))
	for i, k := range sortedKeys {
		hv[i] = fmt.Sprintf(formatStr, k, h[k])
	}
	return fmt.Sprintf("{%s}", strings.Join(hv, ","))
}
