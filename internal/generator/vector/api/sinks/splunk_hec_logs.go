package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type SplunkHecLogs struct {
	Type          types.SinkType `json:"type,omitempty"`
	Inputs        []string       `json:"inputs,omitempty"`
	Endpoint      string         `json:"endpoint,omitempty"`
	DefaultToken  string         `json:"default_token,omitempty"`
	Index         string         `json:"index,omitempty"`
	TimestampKey  string         `json:"timestamp_key,omitempty"`
	IndexedFields []string       `json:"indexed_fields,omitempty"`
	Source        string         `json:"source,omitempty"`
	SourceType    string         `json:"sourcetype,omitempty"`
	HostKey       string         `json:"host_key,omitempty"`
	BaseSink
}

func NewSplunkHecLogs(endpoint string, init func(s *SplunkHecLogs), inputs ...string) (s *SplunkHecLogs) {
	sort.Strings(inputs)
	s = &SplunkHecLogs{
		Type:     types.SinkTypeSplunkHecLogs,
		Inputs:   inputs,
		Endpoint: endpoint,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *SplunkHecLogs) SinkType() types.SinkType {
	return s.Type
}
