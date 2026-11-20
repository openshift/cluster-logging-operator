package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type GcpStackdriverLogs struct {
	Type             types.SinkType `json:"type,omitempty"`
	Inputs           []string       `json:"inputs,omitempty"`
	BillingAccountId string         `json:"billing_account_id,omitempty"`
	CredentialsPath  string         `json:"credentials_path,omitempty"`
	LogId            string         `json:"log_id,omitempty"`
	SeverityKey      string         `json:"severity_key,omitempty"`
	FolderId         string         `json:"folder_id,omitempty"`
	ProjectId        string         `json:"project_id,omitempty"`
	OrganizationId   string         `json:"organization_id,omitempty"`

	// Resource must include 'type'
	Resource map[string]string `json:"resource,omitempty"`
	// TODO: Replace the following with BaseSink?  The public API does not mention
	// compression support but otherwise it is the same.

	Encoding         *Encoding         `json:"encoding,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty"`
	TLS              *transport.TLS    `json:"tls,omitempty"`
}

func NewGcpStackdriverLogs(init func(s *GcpStackdriverLogs), inputs ...string) (s *GcpStackdriverLogs) {
	sort.Strings(inputs)
	s = &GcpStackdriverLogs{
		Type:   types.SinkTypeGcpStackdriverLogs,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *GcpStackdriverLogs) SinkType() types.SinkType {
	return s.Type
}
