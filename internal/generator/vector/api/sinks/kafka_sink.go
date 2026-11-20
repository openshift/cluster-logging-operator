package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type Kafka struct {
	Type               types.SinkType        `json:"type,omitempty"`
	Inputs             []string              `json:"inputs,omitempty"`
	BootstrapServers   string                `json:"bootstrap_servers,omitempty"`
	Topic              string                `json:"topic,omitempty"`
	Compression        CompressionType       `json:"compression,omitempty"`
	HealthCheck        *HealthCheck          `json:"healthcheck,omitempty"`
	Encoding           *Encoding             `json:"encoding,omitempty"`
	Acknowledgements   *Acknowledgements     `json:"acknowledgements,omitempty"`
	Batch              *Batch                `json:"batch,omitempty"`
	Buffer             *Buffer               `json:"buffer,omitempty"`
	Sasl               *Sasl                 `json:"sasl,omitempty"`
	TLS                *transport.TlsEnabled `json:"tls,omitempty"`
	LibrdKafka_Options map[string]string     `json:"librdkafka_options,omitempty"`
}

func NewKafka(init func(s *Kafka), inputs ...string) (s *Kafka) {
	sort.Strings(inputs)
	s = &Kafka{
		Type:   types.SinkTypeKafka,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *Kafka) SinkType() types.SinkType {
	return s.Type
}

type Sasl struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Mechanism string `json:"mechanism,omitempty"`
}
