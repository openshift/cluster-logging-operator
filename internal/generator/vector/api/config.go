package api

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
)

// Config represents a configuration for vector
type Config struct {

	// Transforms is the set of transform ids to transform configurations
	Transforms map[string]interface{} `json:"transforms" yaml:"transforms" toml:"transforms"`
}

// Name is a deprecated method to adapt to the existing generator framework
func (c Config) Name() string {
	return "config"
}

// Template is a deprecated method to adapt to the existing generator framework
func (c Config) Template() string {
	return `{{define "` + c.Name() + `" -}}
{{ if ne "" .String }}
{{.}}
{{end}}
{{end}}`
}

func (c Config) String() string {
	return strings.ReplaceAll(toml.MustMarshal(c), "[transforms]", "")
}
