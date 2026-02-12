package sinks

type AwsCloudwatchLogs struct {
	Type       SinkType `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs     []string `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Region     string   `json:"region,omitempty" yaml:"region,omitempty" toml:"region,omitempty"`
	GroupName  string   `json:"group_name,omitempty" yaml:"group_name,omitempty" toml:"group_name,omitempty"`
	StreamName string   `json:"stream_name,omitempty" yaml:"stream_name,omitempty" toml:"stream_name,omitempty"`
	Endpoint   string   `json:"endpoint,omitempty" yaml:"endpoint,omitempty" toml:"endpoint,omitempty"`

	BaseSink

	Auth *AwsAuth `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`

	HealthCheck HealthCheck `json:"healthcheck" yaml:"healthcheck" toml:"healthcheck"`
}

type HealthCheck struct {
	Enabled bool `json:"enabled" yaml:"enabled" toml:"enabled"`
}

func NewAwsCloudwatchLogs(init func(s *AwsCloudwatchLogs), inputs ...string) (s *AwsCloudwatchLogs) {
	s = &AwsCloudwatchLogs{
		Type:   SinkTypeAwsCloudwatchLogs,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}
