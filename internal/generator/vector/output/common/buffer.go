package common

import "github.com/openshift/cluster-logging-operator/internal/generator/helpers"

const (
	BufferWhenFullBlock      = "block"
	BufferWhenFullDropNewest = "drop_newest"
)

type Buffer struct {
	ComponentID string
	WhenFull    helpers.OptionalPair
	MaxEvents   helpers.OptionalPair
}

func NewBuffer(id string, s ConfigStrategy) Buffer {
	b := Buffer{
		ComponentID: id,
		WhenFull:    helpers.NewOptionalPair("when_full", nil),
		MaxEvents:   helpers.NewOptionalPair("max_events", nil),
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
		b.WhenFull.String()+
		b.MaxEvents.String() == ""
}

func (b Buffer) Template() string {
	if b.isEmpty() {
		return `{{define "` + b.Name() + `" -}}{{end}}`
	}
	return `{{define "` + b.Name() + `" -}}
[sinks.{{.ComponentID}}.buffer]
{{.WhenFull}}
{{.MaxEvents}}
{{end}}`
}
