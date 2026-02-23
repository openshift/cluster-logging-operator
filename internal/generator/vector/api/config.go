package api

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
)

// Config represents a configuration for vector
type Config struct {
	Global

	// Api is the set of API keys to values
	Api *Api `json:"api,omitempty" yaml:"api,omitempty" toml:"api,omitempty"`

	// Secrets is the set of secret ids to secret configurations
	Secret map[string]interface{} `json:"secret,omitempty" yaml:"secret,omitempty" toml:"secret,omitempty"`

	// Sources is the set of source ids to source configurations
	Sources map[string]interface{} `json:"sources,omitempty" yaml:"sources,omitempty" toml:"sources,omitempty"`

	// Transforms is the set of transform ids to transform configurations
	Transforms map[string]interface{} `json:"transforms,omitempty" yaml:"transforms,omitempty" toml:"transforms,omitempty"`

	Sinks map[string]interface{} `json:"sinks,omitempty" yaml:"sinks,omitempty" toml:"sinks,omitempty"`
}

type CodecType string

const (
	CodecTypeJSON CodecType = "json"
)

func NewConfig(init func(*Config)) *Config {
	c := &Config{
		Secret:     make(map[string]interface{}),
		Sources:    make(map[string]interface{}),
		Transforms: make(map[string]interface{}),
		Sinks:      make(map[string]interface{}),
	}
	if init != nil {
		init(c)
	}
	return c
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
	out := strings.ReplaceAll(toml.MustMarshal(c), "[transforms]", "")
	out = strings.ReplaceAll(out, "[sources]", "")
	out = strings.ReplaceAll(out, "[sinks]", "")
	out = strings.ReplaceAll(out, "[secret]", "")
	return out
}
