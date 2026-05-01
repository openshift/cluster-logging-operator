package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type PrometheusExporterAuthStrategy string

const (
	PrometheusExporterAuthStrategySar PrometheusExporterAuthStrategy = "sar"
)

// PrometheusExporterAuth represents authentication configuration for the prometheus exporter
type PrometheusExporterAuth struct {
	Strategy      PrometheusExporterAuthStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty" toml:"strategy,omitempty"`
	Path          string                         `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	Resource      string                         `json:"resource,omitempty" yaml:"resource,omitempty" toml:"resource,omitempty"`
	Verb          string                         `json:"verb,omitempty" yaml:"verb,omitempty" toml:"verb,omitempty"`
	ResourceGroup string                         `json:"resource_group,omitempty" yaml:"resource_group,omitempty" toml:"resource_group,omitempty"`
	Namespace     string                         `json:"namespace,omitempty" yaml:"namespace,omitempty" toml:"namespace,omitempty"`
	User          string                         `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	Groups        []string                       `json:"groups,omitempty" yaml:"groups,omitempty" toml:"groups,omitempty"`
}

type PrometheusExporter struct {
	Type             types.SinkType              `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string                    `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Address          string                      `json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	DefaultNamespace string                      `json:"default_namespace,omitempty" yaml:"default_namespace,omitempty" toml:"default_namespace,omitempty"`
	TLS              *transport.TlsEnabled       `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
	Auth             *PrometheusExporterAuth     `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
}

func (p PrometheusExporter) SinkType() types.SinkType {
	return p.Type
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
