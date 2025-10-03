package api

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Component interface {
	GetID() string
}

type Sinks map[string]Component

type Config struct {
	Sinks   Sinks
	Sources map[string]Component `json:"sources" yaml:"sources" toml:"sources"`
}

func NewConfig() *Config {
	return &Config{
		Sinks:   make(Sinks),
		Sources: make(map[string]Component),
	}
}

func (c *Config) AddSink(s Component) *Config {
	c.Sinks[s.GetID()] = s
	return c
}

func (c *Config) Name() string {
	return "config"
}

func (c *Config) Template() string {
	return `{{define "` + c.Name() + `" -}}{{.}}{{end}}}`
}

func (c *Config) String() string {
	out, err := c.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (c *Config) MarshalText() (text []byte, err error) {
	out := new(bytes.Buffer)
	encoder := toml.NewEncoder(out)
	encoder.Indent = ""
	for k, sink := range c.Sinks {
		if _, err := out.WriteString(fmt.Sprintf("[sinks.%s]\n", k)); err != nil {
			return nil, err
		}
		if err := encoder.Encode(sink); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}
