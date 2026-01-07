package remap

import (
	vectorapi "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
)

type Remap struct {
	id string
	// Type is required to be 'remop'
	Type string `json:"type" yaml:"type" toml:"type"`

	// Inputs is the IDs of the components feeding into this component
	Inputs []string `json:"inputs" yaml:"inputs" toml:"inputs"`

	// Source is the VRL script used for the remap transformation
	Source string `json:"source" yaml:"source" toml:"source" multiline:"true" literal:"true"`
}

func New(id, source string, inputs ...string) *Remap {
	return &Remap{
		id:     id,
		Type:   "remap",
		Inputs: inputs,
		Source: source,
	}
}

// Name is a deprecated method to adapt to the existing generator framework
func (r Remap) Name() string {
	return "remap"
}

// Template is a deprecated method to adapt to the existing generator framework
func (r Remap) Template() string {
	return `{{define "` + r.Name() + `" -}}
{{ if ne "" .String }}
{{.}}
{{end}}
{{end}}`
}

func (r Remap) String() string {
	c := vectorapi.Config{
		Transforms: map[string]interface{}{
			r.id: r,
		},
	}
	return c.String()
}
