package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type PrometheusExporter struct {
	Type             types.SinkType        `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string              `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Address          string                `json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	DefaultNamespace string                `json:"default_namespace,omitempty" yaml:"default_namespace,omitempty" toml:"default_namespace,omitempty"`
	TLS              *transport.TlsEnabled `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func (p PrometheusExporter) SinkType() string {
	return string(p.Type)
}

func NewPrometheusExporter(address string, init func(s *PrometheusExporter), inputs ...string) (s *PrometheusExporter) {
	sort.Strings(inputs)
	s = &PrometheusExporter{
		Type:    types.SinkTypePrometheusExporter,
		Inputs:  inputs,
		Address: address,
	}
	if init != nil {
		init(s)
	}
	return s
}
