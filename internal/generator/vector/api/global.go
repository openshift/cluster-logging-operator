package api

// Global are the root level global configuration
type Global struct {
	ExpireMetricsSec uint   `json:"expire_metrics_secs,omitempty"`
	DataDir          string `json:"data_dir,omitempty"`
}
