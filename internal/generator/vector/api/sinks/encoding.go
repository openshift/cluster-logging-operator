package sinks

import (
	utilstoml "github.com/openshift/cluster-logging-operator/internal/utils/toml"
)

const (
	CodecJSON              = "json"
	TimeStampFormatRFC3339 = "rfc3339"
)

// Encoding configures the encoding of events.
type Encoding struct {
	ID              string   `toml:"-"`
	Codec           string   `json:"codec,omitempty" yaml:"codec,omitempty" toml:"codec,omitempty"`
	ExceptFields    []string `json:"except_fields,omitempty" yaml:"except_fields,omitempty" toml:"except_fields,omitempty"`
	TimeStampFormat string   `json:"timestamp_format,omitempty" yaml:"timestamp_format,omitempty" toml:"timestamp_format,omitempty"`
}

// NewEncoding initializes an encoding with the given codec is capable of taking additional
// initializers to set other fields as needed
func NewEncoding(id string, codec string, inits ...func(*Encoding)) Encoding {
	e := Encoding{
		ID:           id,
		Codec:        codec,
		ExceptFields: []string{"_internal"},
	}
	for _, init := range inits {
		init(&e)
	}
	return e
}

func (e Encoding) Name() string {
	return "encoding"
}

func (e Encoding) Template() string {
	return `{{define "` + e.Name() + `" -}}
{{ if ne "" .String }}
[sinks.{{.ID}}.encoding]
{{.}}
{{end}}
{{end}}`
}

func (e Encoding) String() string {
	return utilstoml.MustMarshal(e)
}
