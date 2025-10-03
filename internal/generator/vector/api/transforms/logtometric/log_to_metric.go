package logtometric

import (
	"fmt"
	"sort"
	"strings"
)

// LogToMetric is the configuration for the log_to_metric transform
type LogToMetric struct {
	// Type is required to be 'log_to_metric'
	Type string `json:"type" yaml:"type" toml:"type"`

	// Inputs is the IDs of the components feeding into this component
	Inputs []string `json:"inputs" yaml:"inputs" toml:"inputs"`

	// Metrics is the spec for the Metrics being exposed
	Metrics []Metric `json:"metrics" yaml:"metrics" toml:"metrics"`
}

func New(name string, metricsType MetricsType, tags Tags, inputs ...string) *LogToMetric {
	return &LogToMetric{
		Type:   "log_to_metric",
		Inputs: inputs,
		Metrics: []Metric{{
			MetricName: name,
			Namespace:  "logcollector",
			Field:      "message",
			Kind:       MetricsKindIncremental,
			Type:       metricsType,
			Tags:       tags,
		}},
	}
}

type MetricsKind string

type MetricsType string

const (
	MetricsKindAbsolute MetricsKind = "absolute"

	// MetricsKindIncremental default if not defined
	MetricsKindIncremental MetricsKind = "incremental"

	MetricsTypeCounter MetricsType = "counter"
)

type Metric struct {
	Field string `json:"field" yaml:"field" toml:"field"`

	MetricName string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`

	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty" toml:"namespace,omitempty"`

	Kind MetricsKind `json:"kind,omitempty" yaml:"kind,omitempty" toml:"kind,omitempty"`

	Type MetricsType `json:"type" yaml:"type" toml:"type"`

	// Tags optional tags (or labels) to apply to the metric
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"  toml:"tags,omitempty,multiline"`
}

// Tags optional tags (or labels) to apply to the metric
type Tags map[string]string

func (t Tags) AddAll(tags map[string]string) {
	if tags == nil {
		return
	}
	for k, v := range tags {
		t[k] = v
	}
}

func (t Tags) MarshalTOML() ([]byte, error) {
	tags := []string{}
	for key, value := range t {
		tags = append(tags, fmt.Sprintf("%s = %q", key, value))
	}
	sort.Strings(tags)
	result := strings.Join(tags, ", ")
	return []byte(fmt.Sprintf("{%v}", result)), nil
}
