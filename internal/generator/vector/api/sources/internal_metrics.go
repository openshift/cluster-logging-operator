package sources

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type InternalMetrics struct {
	Type                 types.SourceType `json:"type" yaml:"type" toml:"type"`
	ScrapIntervalSeconds uint             `json:"scrape_interval_seconds,omitempty" yaml:"scrape_interval_seconds,omitempty" toml:"scrape_interval_seconds,omitempty"`
}

func (i *InternalMetrics) SourceType() types.SourceType {
	return types.SourceTypeInternalMetrics
}

func NewInternalMetrics() types.Source {
	return &InternalMetrics{
		Type: types.SourceTypeInternalMetrics,
	}
}
