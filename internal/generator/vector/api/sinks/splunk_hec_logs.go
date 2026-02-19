package sinks

type SplunkHecLogs struct {
	Type          SinkType `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs        []string `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Endpoint      string   `json:"endpoint,omitempty" yaml:"endpoint,omitempty" toml:"endpoint,omitempty"`
	DefaultToken  string   `json:"default_token,omitempty" yaml:"default_token,omitempty" toml:"default_token,omitempty"`
	Index         string   `json:"index,omitempty" yaml:"index,omitempty" toml:"index,omitempty"`
	TimestampKey  string   `json:"timestamp_key,omitempty" yaml:"timestamp_key,omitempty" toml:"timestamp_key,omitempty"`
	IndexedFields []string `json:"indexed_fields,omitempty" yaml:"indexed_fields,omitempty" toml:"indexed_fields,omitempty"`
	Source        string   `json:"source,omitempty" yaml:"source,omitempty" toml:"source,omitempty"`
	SourceType    string   `json:"sourcetype,omitempty" yaml:"sourcetype,omitempty" toml:"sourcetype,omitempty"`
	HostKey       string   `json:"host_key,omitempty" yaml:"host_key,omitempty" toml:"host_key,omitempty"`
	BaseSink
}

func NewSplunkHecLogs(endpoint string, init func(s *SplunkHecLogs), inputs ...string) (s *SplunkHecLogs) {
	s = &SplunkHecLogs{
		Type:     SinkTypeHecLogs,
		Inputs:   inputs,
		Endpoint: endpoint,
	}
	if init != nil {
		init(s)
	}
	return s
}
