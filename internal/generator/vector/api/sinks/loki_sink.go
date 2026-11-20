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
	Type             types.SinkType       `json:"type,omitempty"`
	Inputs           []string             `json:"inputs,omitempty"`
	Endpoint         string               `json:"endpoint,omitempty"`
	OutOfOrderAction LokiOutOfOrderAction `json:"out_of_order_action,omitempty"`
	TenantId         string               `json:"tenant_id,omitempty"`
	BaseSink
	HealthCheck *HealthCheck      `json:"healthcheck,omitempty"`
	Auth        *HttpAuth         `json:"auth,omitempty"`
	Proxy       *Proxy            `json:"proxy,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
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

func (s *Loki) SinkType() types.SinkType {
	return s.Type
}
