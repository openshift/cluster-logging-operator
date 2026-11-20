package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type AwsCloudwatchLogs struct {
	Type       types.SinkType `json:"type,omitempty"`
	Inputs     []string       `json:"inputs,omitempty"`
	Region     string         `json:"region,omitempty"`
	GroupName  string         `json:"group_name,omitempty"`
	StreamName string         `json:"stream_name,omitempty"`
	Endpoint   string         `json:"endpoint,omitempty"`

	BaseSink

	Auth *AwsAuth `json:"auth,omitempty"`

	HealthCheck HealthCheck `json:"healthcheck"`
}

type HealthCheck struct {
	Enabled bool `json:"enabled"`
}

func NewAwsCloudwatchLogs(init func(s *AwsCloudwatchLogs), inputs ...string) (s *AwsCloudwatchLogs) {
	sort.Strings(inputs)
	s = &AwsCloudwatchLogs{
		Type:   types.SinkTypeAwsCloudwatchLogs,
		Inputs: inputs,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *AwsCloudwatchLogs) SinkType() types.SinkType {
	return s.Type
}
