package sinks

type AwsS3 struct {
	Type      SinkType `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs    []string `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Region    string   `json:"region,omitempty" yaml:"region,omitempty" toml:"region,omitempty"`
	Bucket    string   `json:"bucket,omitempty" yaml:"bucket,omitempty" toml:"bucket,omitempty"`
	KeyPrefix string   `json:"key_prefix,omitempty" yaml:"key_prefix,omitempty" toml:"key_prefix,omitempty"`
	Endpoint  string   `json:"endpoint,omitempty" yaml:"endpoint,omitempty" toml:"endpoint,omitempty"`

	BaseSink

	Auth *AwsAuth `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`

	HealthCheck HealthCheck `json:"healthcheck" yaml:"healthcheck" toml:"healthcheck"`
}

func NewAwsS3(init func(s *AwsS3), inputs ...string) (s *AwsS3) {
	s = &AwsS3{
		Type:   SinkTypeAwsS3,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}
