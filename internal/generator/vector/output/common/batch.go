package common

import (
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

type Batch struct {
	ComponentID string
	MaxBytes    helpers.OptionalPair
	MaxEvents   helpers.OptionalPair
	TimeoutSec  helpers.OptionalPair
}

// NewApiBatch returns the batch tuning for an output or nil when nothing varies
// from the defaults
func NewApiBatch(t observability.TunableOutput) *sinks.Batch {
	maxBytes := t.GetTuning().MaxWrite
	if maxBytes != nil && !maxBytes.IsZero() {
		return &sinks.Batch{
			MaxBytes: uint64(maxBytes.Value()),
		}
	}
	return nil
}

func NewBatch(id string, s ConfigStrategy) Batch {
	b := Batch{
		ComponentID: id,
		MaxBytes:    helpers.NewOptionalPair("max_bytes", nil),
		MaxEvents:   helpers.NewOptionalPair("max_events", nil),
		TimeoutSec:  helpers.NewOptionalPair("timeout_sec", nil),
	}
	if s != nil {
		b = s.VisitBatch(b)
	}
	return b
}

func (b Batch) Name() string {
	return "batch"
}

func (b Batch) isEmpty() bool {
	return b.MaxEvents.String()+
		b.MaxBytes.String()+
		b.TimeoutSec.String() == ""
}

func (b Batch) Template() string {
	if b.isEmpty() {
		return `{{define "` + b.Name() + `" -}}{{end}}`
	}
	return `{{define "` + b.Name() + `" -}}
[sinks.{{.ComponentID}}.batch]
{{.MaxBytes}}
{{.MaxEvents}}
{{.TimeoutSec}}
{{end}}`
}
