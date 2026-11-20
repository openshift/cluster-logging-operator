package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type OpenTelemetry struct {
	Type     types.SinkType         `json:"type,omitempty"`
	Inputs   []string               `json:"inputs,omitempty"`
	Protocol *OpenTelemetryProtocol `json:"protocol,omitempty"`
	Batch    *Batch                 `json:"batch,omitempty"`
	Buffer   *Buffer                `json:"buffer,omitempty"`
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
	URI           string          `json:"uri,omitempty"`
	Type          string          `json:"type,omitempty"`
	Method        MethodType      `json:"method,omitempty"`
	PayloadPrefix string          `json:"payload_prefix,omitempty"`
	PayloadSuffix string          `json:"payload_suffix,omitempty"`
	Compression   CompressionType `json:"compression,omitempty"`
	TLS           *transport.TLS  `json:"tls,omitempty"`
	Encoding      *Encoding       `json:"encoding,omitempty"`
	Auth          *HttpAuth       `json:"auth,omitempty"`
	Request       *Request        `json:"request,omitempty"`
}

func (s *OpenTelemetry) SinkType() types.SinkType {
	return s.Type
}
