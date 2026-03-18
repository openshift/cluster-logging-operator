package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type OpenTelemetry struct {
	Type     types.SinkType         `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs   []string               `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Protocol *OpenTelemetryProtocol `json:"protocol,omitempty" yaml:"protocol,omitempty" toml:"protocol,omitempty"`
	Batch    *Batch                 `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer   *Buffer                `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
}

func NewOpenTelemetry(uri string, init func(telemetry *OpenTelemetry), inputs ...string) *OpenTelemetry {
	sort.Strings(inputs)
	s := &OpenTelemetry{
		Type:   types.SinkTypeOpenTelemetry,
		Inputs: inputs,
		Protocol: &OpenTelemetryProtocol{
			URI: uri,
		},
	}
	if init != nil {
		init(s)
	}
	return s
}

type OpenTelemetryProtocol struct {
	URI           string          `json:"uri,omitempty" yaml:"uri,omitempty" toml:"uri,omitempty"`
	Type          string          `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Method        MethodType      `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	PayloadPrefix string          `json:"payload_prefix,omitempty" yaml:"payload_prefix,omitempty" toml:"payload_prefix,omitempty"`
	PayloadSuffix string          `json:"payload_suffix,omitempty" yaml:"payload_suffix,omitempty" toml:"payload_suffix,omitempty"`
	TLS           *transport.TLS  `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
	Compression   CompressionType `json:"compression,omitempty" yaml:"compression,omitempty" toml:"compression,omitempty"`
	Encoding      *Encoding       `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Auth          *HttpAuth       `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	Request       *Request        `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
}

func (s *OpenTelemetry) SinkType() types.SinkType {
	return s.Type
}
