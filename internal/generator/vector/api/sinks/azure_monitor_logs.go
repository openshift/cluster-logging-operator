package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type AzureMonitorLogs struct {
	Type             types.SinkType    `json:"type,omitempty"`
	Inputs           []string          `json:"inputs,omitempty"`
	AzureResourceId  string            `json:"azure_resource_id,omitempty"`
	CustomerId       string            `json:"customer_id,omitempty"`
	Host             string            `json:"host,omitempty"`
	LogType          string            `json:"log_type,omitempty"`
	SharedKey        string            `json:"shared_key,omitempty"`
	Encoding         *Encoding         `json:"encoding,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty"`
	TLS              *transport.TLS    `json:"tls,omitempty"`
}

func NewAzureMonitorLogs(init func(s *AzureMonitorLogs), inputs ...string) (s *AzureMonitorLogs) {
	sort.Strings(inputs)
	s = &AzureMonitorLogs{
		Type:   types.SinkTypeAzureMonitorLogs,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *AzureMonitorLogs) SinkType() types.SinkType {
	return s.Type
}
