package sinks

import utilstoml "github.com/openshift/cluster-logging-operator/internal/utils/toml"

// Batch configures the batching behavior.
type Batch struct {
	ID          string  `toml:"-"`
	MaxBytes    uint    `json:"max_bytes,omitempty" yaml:"max_bytes,omitempty" toml:"max_bytes,omitempty,omitzero"`
	MaxEvents   uint    `json:"max_events,omitempty" yaml:"max_events,omitempty" toml:"max_events,omitempty,omitzero"`
	TimeoutSecs float64 `json:"timeout_secs,omitempty" yaml:"timeout_secs,omitempty" toml:"timeout_secs,omitempty,omitzero"`
}

func NewBatch(id string, maxBytes, maxEvents uint, timeoutSecs float64) Batch {
	e := Batch{
		ID:          id,
		MaxBytes:    maxBytes,
		MaxEvents:   maxEvents,
		TimeoutSecs: timeoutSecs,
	}
	return e
}

func (b Batch) Name() string {
	return "encoding"
}

func (b Batch) Template() string {
	return `{{define "` + b.Name() + `" -}}
{{ if ne "" .String }}
[sinks.{{.ID}}.batch]
{{.}}
{{end}}
{{end}}`
}

func (b Batch) String() string {
	return utilstoml.MustMarshal(b)
}
