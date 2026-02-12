package common

import (
	"github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

const (
	BufferWhenFullBlock      = "block"
	BufferWhenFullDropNewest = "drop_newest"
	minBufferSize            = 268435488
)

type Buffer struct {
	ComponentID string

	Type      helpers.OptionalPair
	WhenFull  helpers.OptionalPair
	MaxEvents helpers.OptionalPair
	MaxSize   helpers.OptionalPair
}

// NewApiBuffer returns the buffer tuning for an output or nil when nothing varies
// from the defaults
func NewApiBuffer(t observability.TunableOutput) *sinks.Buffer {
	switch t.GetTuning().DeliveryMode {
	case v1.DeliveryModeAtLeastOnce:
		return &sinks.Buffer{
			WhenFull: sinks.BufferWhenFullBlock,
			Type:     sinks.BufferTypeDisk,
			MaxSize:  minBufferSize,
		}
	case v1.DeliveryModeAtMostOnce:
		return &sinks.Buffer{
			WhenFull: sinks.BufferWhenFullDropNewest,
		}
	}
	return nil
}

func NewBuffer(id string, s ConfigStrategy) Buffer {
	b := Buffer{
		ComponentID: id,
		Type:        helpers.NewOptionalPair("type", nil),
		WhenFull:    helpers.NewOptionalPair("when_full", nil),
		MaxEvents:   helpers.NewOptionalPair("max_events", nil),
		MaxSize:     helpers.NewOptionalPair("max_size", nil),
	}
	if s != nil {
		b = s.VisitBuffer(b)
	}
	return b
}

func (b Buffer) Name() string {
	return "buffer"
}

func (b Buffer) isEmpty() bool {
	return b.MaxEvents.String()+
		b.Type.String()+
		b.WhenFull.String()+
		b.MaxSize.String()+
		b.MaxEvents.String() == ""
}

func (b Buffer) Template() string {
	if b.isEmpty() {
		return `{{define "` + b.Name() + `" -}}{{end}}`
	}
	return `{{define "` + b.Name() + `" -}}
[sinks.{{.ComponentID}}.buffer]
{{.Type}}
{{.WhenFull}}
{{.MaxEvents}}
{{.MaxSize}}
{{end}}`
}
