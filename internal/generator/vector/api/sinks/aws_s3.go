package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type AwsS3 struct {
	Type      types.SinkType `json:"type,omitempty"`
	Inputs    []string       `json:"inputs,omitempty"`
	Region    string         `json:"region,omitempty"`
	Bucket    string         `json:"bucket,omitempty"`
	KeyPrefix string         `json:"key_prefix,omitempty"`
	Endpoint  string         `json:"endpoint,omitempty"`

	BaseSink

	Auth *AwsAuth `json:"auth,omitempty"`

	HealthCheck HealthCheck `json:"healthcheck"`
}

func NewAwsS3(init func(s *AwsS3), inputs ...string) (s *AwsS3) {
	sort.Strings(inputs)
	s = &AwsS3{
		Type:   types.SinkTypeAwsS3,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *AwsS3) SinkType() types.SinkType {
	return s.Type
}
