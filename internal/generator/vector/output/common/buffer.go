package common

import "github.com/openshift/cluster-logging-operator/internal/generator/helpers"

const (
	//BufferMinSizeBytes is the minimal buffer required by vector for disk buffers
	BufferMinSizeBytes       = 268435488
	BufferTypeDisk           = "disk"
	BufferWhenFullBlock      = "block"
	BufferWhenFullDropNewest = "drop_newest"
)

type Buffer struct {
	ComponentID string

	Type      helpers.OptionalPair
	WhenFull  helpers.OptionalPair
	MaxEvents helpers.OptionalPair
	MaxSize   helpers.OptionalPair
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
