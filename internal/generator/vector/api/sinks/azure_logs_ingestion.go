package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type AzureLogsIngestion struct {
	Type           types.SinkType `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs         []string       `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Endpoint       string         `json:"endpoint,omitempty" yaml:"endpoint,omitempty" toml:"endpoint,omitempty"`
	DcrImmutableId string         `json:"dcr_immutable_id,omitempty" yaml:"dcr_immutable_id,omitempty" toml:"dcr_immutable_id,omitempty"`
	StreamName     string         `json:"stream_name,omitempty" yaml:"stream_name,omitempty" toml:"stream_name,omitempty"`
	TokenScope     string         `json:"token_scope,omitempty" yaml:"token_scope,omitempty" toml:"token_scope,omitempty"`
	TimestampField string         `json:"timestamp_field,omitempty" yaml:"timestamp_field,omitempty" toml:"timestamp_field,omitempty"`

	Auth *AzureLogsIngestionAuth `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`

	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Encoding         *Encoding         `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Batch            *Batch            `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
	TLS              *transport.TLS    `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

type AzureLogsIngestionAuth struct {
	AzureCredentialKind string `json:"azure_credential_kind,omitempty" yaml:"azure_credential_kind,omitempty" toml:"azure_credential_kind,omitempty"`

	// Specific to Client Secret Auth
	AzureTenantId     string `json:"azure_tenant_id,omitempty" yaml:"azure_tenant_id,omitempty" toml:"azure_tenant_id,omitempty"`
	AzureClientId     string `json:"azure_client_id,omitempty" yaml:"azure_client_id,omitempty" toml:"azure_client_id,omitempty"`
	AzureClientSecret string `json:"azure_client_secret,omitempty" yaml:"azure_client_secret,omitempty" toml:"azure_client_secret,omitempty"`

	// Specific to Workload Identity Auth
	TenantId      string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" toml:"tenant_id,omitempty"`
	ClientId      string `json:"client_id,omitempty" yaml:"client_id,omitempty" toml:"client_id,omitempty"`
	TokenFilePath string `json:"token_file_path,omitempty" yaml:"token_file_path,omitempty" toml:"token_file_path,omitempty"`
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
