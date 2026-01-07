package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type LokiOutOfOrderAction string

const (
	LokiOutOfOrderActionAccept LokiOutOfOrderAction = "accept"
)

type Loki struct {
	Type             types.SinkType       `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string             `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Endpoint         string               `json:"endpoint,omitempty" yaml:"endpoint,omitempty" toml:"endpoint,omitempty"`
	OutOfOrderAction LokiOutOfOrderAction `json:"out_of_order_action,omitempty" yaml:"out_of_order_action,omitempty" toml:"out_of_order_action,omitempty"`
	TenantId         string               `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" toml:"tenant_id,omitempty"`
	HealthCheck      *HealthCheck         `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty" toml:"healthcheck,omitempty"`
	Auth             *HttpAuth            `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	BaseSink
	Proxy  *Proxy            `json:"proxy,omitempty" yaml:"proxy,omitempty" toml:"proxy,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`
}

func NewLoki(endpoint string, init func(s *Loki), inputs ...string) (s *Loki) {
	sort.Strings(inputs)
	s = &Loki{
		Type:     types.SinkTypeLoki,
		Inputs:   inputs,
		Endpoint: endpoint,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *Loki) SinkType() string {
	return string(s.Type)
}
