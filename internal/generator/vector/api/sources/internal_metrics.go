package sources

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type InternalMetrics struct {
	Type               types.SourceType `json:"type"`
	ScrapeIntervalSecs uint             `json:"scrape_interval_secs,omitempty"`
}

func (i *InternalMetrics) SourceType() types.SourceType {
	return types.SourceTypeInternalMetrics
}

func NewInternalMetrics() types.Source {
	return &InternalMetrics{
		Type: types.SourceTypeInternalMetrics,
	}
}
