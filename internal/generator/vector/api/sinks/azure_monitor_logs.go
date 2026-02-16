package sinks

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"

type AzureMonitorLogs struct {
	Type             SinkType          `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string          `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	AzureResourceId  string            `json:"azure_resource_id,omitempty" yaml:"azure_resource_id,omitempty" toml:"azure_resource_id,omitempty"`
	CustomerId       string            `json:"customer_id,omitempty" yaml:"customer_id,omitempty" toml:"customer_id,omitempty"`
	Host             string            `json:"host,omitempty" yaml:"host,omitempty" toml:"host,omitempty"`
	LogType          string            `json:"log_type,omitempty" yaml:"log_type,omitempty" toml:"log_type,omitempty"`
	SharedKey        string            `json:"shared_key,omitempty" yaml:"shared_key,omitempty" toml:"shared_key,omitempty"`
	Encoding         *Encoding         `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
	TLS              *api.TLS          `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func NewAzureMonitorLogs(init func(s *AzureMonitorLogs), inputs ...string) (s *AzureMonitorLogs) {
	s = &AzureMonitorLogs{
		Type:   SinkTypeAzureMonitorLogs,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}
