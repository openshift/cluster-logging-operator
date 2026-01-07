package api

// Global are the root level global configuration
type Global struct {
	DataDir          string `json:"data_dir,omitempty" yaml:"data_dir,omitempty" toml:"data_dir,omitempty"`
	ExpireMetricsSec uint   `json:"expire_metrics_secs,omitempty" yaml:"expire_metrics_secs,omitempty" toml:"expire_metrics_secs,omitempty"`
}
