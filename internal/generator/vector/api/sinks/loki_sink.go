package sinks

type LokiOutOfOrderAction string

const (
	LokiOutOfOrderActionAccept LokiOutOfOrderAction = "accept"
)

type Loki struct {
	Type             SinkType             `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string             `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Endpoint         string               `json:"endpoint,omitempty" yaml:"endpoint,omitempty" toml:"endpoint,omitempty"`
	OutOfOrderAction LokiOutOfOrderAction `json:"out_of_order_action,omitempty" yaml:"out_of_order_action,omitempty" toml:"out_of_order_action,omitempty"`
	TenantId         string               `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" toml:"tenant_id,omitempty"`
	HealthCheck      HealthCheck          `json:"healthcheck" yaml:"healthcheck" toml:"healthcheck"`
	Auth             *HttpAuth            `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	BaseSink
	Proxy  *Proxy            `json:"proxy,omitempty" yaml:"proxy,omitempty" toml:"proxy,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`
}

func NewLoki(endpoint string, init func(s *Loki), inputs ...string) (s *Loki) {
	s = &Loki{
		Type:     SinkTypeLoki,
		Inputs:   inputs,
		Endpoint: endpoint,
	}
	if init != nil {
		init(s)
	}
	return s
}
