package sources

type InternalMetrics struct {
	// Type is required to be 'internal_metrics'
	Type SourceType `json:"type" yaml:"type" toml:"type"`

	ScrapIntervalSeconds uint `json:"scrap_interval_seconds,omitempty" yaml:"scrap_interval_seconds,omitempty" toml:"scrap_interval_seconds,omitempty"`
}

func NewInternalMetrics() *InternalMetrics {
	return &InternalMetrics{
		Type: SourceTypeInternalMetrics,
	}
}
