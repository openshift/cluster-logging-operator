package types

// TlsClientConfig contains TLS client configuration
type TlsClientConfig struct {
	Insecure           bool   `yaml:"insecure,omitempty"`             // Disable TLS verification
	CAFile             string `yaml:"ca_file,omitempty"`              // Path to CA certificate file
	CertFile           string `yaml:"cert_file,omitempty"`            // Path to client certificate file
	KeyFile            string `yaml:"key_file,omitempty"`             // Path to client key file
	ServerNameOverride string `yaml:"server_name_override,omitempty"` // Override server name for TLS verification
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"` // Skip TLS verification (deprecated, use Insecure)
	MinVersion         string `yaml:"min_version,omitempty"`          // Minimum TLS version (e.g., "1.2")
	MaxVersion         string `yaml:"max_version,omitempty"`          // Maximum TLS version (e.g., "1.3")
}

// QueueSettings defines configuration for the exporter's sending queue
type QueueSettings struct {
	Enabled      bool `yaml:"enabled,omitempty"`       // Whether to enable the queue (default: true)
	NumConsumers int  `yaml:"num_consumers,omitempty"` // Number of consumers for the queue (default: 10)
	QueueSize    int  `yaml:"queue_size,omitempty"`    // Size of the queue (default: 1000)
}

// RetrySettings defines configuration for retry behavior on export failure
type RetrySettings struct {
	Enabled         bool   `yaml:"enabled,omitempty"`          // Whether retries are enabled (default: true)
	InitialInterval string `yaml:"initial_interval,omitempty"` // Initial interval for retry backoff (default: "5s")
	MaxInterval     string `yaml:"max_interval,omitempty"`     // Maximum interval for retry backoff (default: "30s")
	MaxElapsedTime  string `yaml:"max_elapsed_time,omitempty"` // Maximum time to retry (default: "5m")
}
