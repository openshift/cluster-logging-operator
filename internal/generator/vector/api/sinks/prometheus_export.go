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
	Strategy      PrometheusExporterAuthStrategy `json:"strategy,omitempty"`
	Path          string                         `json:"path,omitempty"`
	Resource      string                         `json:"resource,omitempty"`
	Verb          string                         `json:"verb,omitempty"`
	ResourceGroup string                         `json:"resource_group,omitempty"`
	Namespace     string                         `json:"namespace,omitempty"`
	User          string                         `json:"user,omitempty"`
	Groups        []string                       `json:"groups,omitempty"`
}

type PrometheusExporter struct {
	Type             types.SinkType          `json:"type,omitempty"`
	Inputs           []string                `json:"inputs,omitempty"`
	Address          string                  `json:"address,omitempty"`
	DefaultNamespace string                  `json:"default_namespace,omitempty"`
	TLS              *transport.TlsEnabled   `json:"tls,omitempty"`
	Auth             *PrometheusExporterAuth `json:"auth,omitempty"`
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
