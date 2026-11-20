package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type AzureLogsIngestion struct {
	Type           types.SinkType `json:"type,omitempty"`
	Inputs         []string       `json:"inputs,omitempty"`
	Endpoint       string         `json:"endpoint,omitempty"`
	DcrImmutableId string         `json:"dcr_immutable_id,omitempty"`
	StreamName     string         `json:"stream_name,omitempty"`
	TokenScope     string         `json:"token_scope,omitempty"`
	TimestampField string         `json:"timestamp_field,omitempty"`

	Auth *AzureLogsIngestionAuth `json:"auth,omitempty"`

	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty"`
	Encoding         *Encoding         `json:"encoding,omitempty"`
	Batch            *Batch            `json:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty"`
	TLS              *transport.TLS    `json:"tls,omitempty"`
}

type AzureLogsIngestionAuth struct {
	AzureCredentialKind string `json:"azure_credential_kind,omitempty"`

	// Specific to Client Secret Auth
	AzureTenantId     string `json:"azure_tenant_id,omitempty"`
	AzureClientId     string `json:"azure_client_id,omitempty"`
	AzureClientSecret string `json:"azure_client_secret,omitempty"`

	// Specific to Workload Identity Auth
	TenantId      string `json:"tenant_id,omitempty"`
	ClientId      string `json:"client_id,omitempty"`
	TokenFilePath string `json:"token_file_path,omitempty"`
}

func NewAzureLogsIngestion(init func(s *AzureLogsIngestion), inputs ...string) (s *AzureLogsIngestion) {
	sort.Strings(inputs)
	s = &AzureLogsIngestion{
		Type:   types.SinkTypeAzureLogsIngestion,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *AzureLogsIngestion) SinkType() types.SinkType {
	return s.Type
}
