package sinks

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"

type Kafka struct {
	Type               SinkType          `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs             []string          `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	BootstrapServers   string            `json:"bootstrap_servers,omitempty" yaml:"bootstrap_servers,omitempty" toml:"bootstrap_servers,omitempty"`
	Topic              string            `json:"topic,omitempty" yaml:"topic,omitempty" toml:"topic,omitempty"`
	Compression        CompressionType   `json:"compression,omitempty" yaml:"compression,omitempty" toml:"compression,omitempty"`
	HealthCheck        *HealthCheck      `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty" toml:"healthcheck,omitempty"`
	Encoding           *Encoding         `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Acknowledgements   *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Batch              *Batch            `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer             *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	Sasl               *Sasl             `json:"sasl,omitempty" yaml:"sasl,omitempty" toml:"sasl,omitempty"`
	TLS                *api.TLS          `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
	LibrdKafka_Options map[string]string `json:"librdkafka_options,omitempty" yaml:"librdkafka_options,omitempty" toml:"librdkafka_options,omitempty"`
}

func NewKafka(init func(s *Kafka), inputs ...string) (s *Kafka) {
	s = &Kafka{
		Type:   SinkTypeKafka,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

type Sasl struct {
	Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`
	Username  string `json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password  string `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
	Mechanism string `json:"mechanism,omitempty" yaml:"mechanism,omitempty" toml:"mechanism,omitempty"`
}
