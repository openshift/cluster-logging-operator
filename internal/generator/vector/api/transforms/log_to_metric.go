package transforms

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

// LogToMetric is the configuration for the log_to_metric transform
type LogToMetric struct {
	// Inputs is the IDs of the components feeding into this component
	Inputs []string `json:"inputs"`

	// Type is required to be 'log_to_metric'
	Type types.TransformType `json:"type"`

	// Metrics is the spec for the Metrics being exposed
	Metrics []Metric `json:"metrics"`
}

func NewLogToMetric(name string, metricsType MetricsType, tags Tags, inputs ...string) *LogToMetric {
	sort.Strings(inputs)
	return &LogToMetric{
		Type:   types.TransformTypeLogToMetric,
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

func (t *LogToMetric) TransformType() types.TransformType {
	return t.Type
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
	Field string `json:"field"`

	Kind MetricsKind `json:"kind,omitempty"`

	MetricName string `json:"name,omitempty"`

	Namespace string `json:"namespace,omitempty"`

	// Tags optional tags (or labels) to apply to the metric
	Tags Tags `json:"tags,omitempty"`

	Type MetricsType `json:"type"`
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
