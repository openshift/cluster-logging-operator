package sinks

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"

type PrometheusExporter struct {
	Type             SinkType `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Address          string   `json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	DefaultNamespace string   `json:"default_namespace,omitempty" yaml:"default_namespace,omitempty" toml:"default_namespace,omitempty"`
	TLS              *api.TLS `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func NewPrometheusExporter(address string, init func(s *PrometheusExporter), inputs ...string) (s *PrometheusExporter) {
	s = &PrometheusExporter{
		Type:    SinkTypePrometheusExporter,
		Inputs:  inputs,
		Address: address,
	}
	if init != nil {
		init(s)
	}
	return s
}
