package sinks

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"

type GcpStackdriverLogs struct {
	Type             SinkType `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	BillingAccountId string   `json:"billing_account_id,omitempty" yaml:"billing_account_id,omitempty" toml:"billing_account_id,omitempty"`
	CredentialsPath  string   `json:"credentials_path,omitempty" yaml:"credentials_path,omitempty" toml:"credentials_path,omitempty"`
	LogId            string   `json:"log_id,omitempty" yaml:"log_id,omitempty" toml:"log_id,omitempty"`
	SeverityKey      string   `json:"severity_key,omitempty" yaml:"severity_key,omitempty" toml:"severity_key,omitempty"`
	FolderId         string   `json:"folder_id,omitempty" yaml:"folder_id,omitempty" toml:"folder_id,omitempty"`
	ProjectId        string   `json:"project_id,omitempty" yaml:"project_id,omitempty" toml:"project_id,omitempty"`
	OrganizationId   string   `json:"organization_id,omitempty" yaml:"organization_id,omitempty" toml:"organization_id,omitempty"`

	Resource *Resource `json:"resource,omitempty" yaml:"resource,omitempty" toml:"resource,omitempty"`
	// TODO: Replace the following with BaseSink?  The public API does not mention
	// compression support but otherwise it is the same.

	Encoding         *Encoding         `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
	TLS              *api.TLS          `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func NewGcpStackdriverLogs(init func(s *GcpStackdriverLogs), inputs ...string) (s *GcpStackdriverLogs) {
	s = &GcpStackdriverLogs{
		Type:   SinkTypeGcpStackdriverLogs,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

type Resource struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	ResourceLabels
}

type ResourceLabels map[string]string
