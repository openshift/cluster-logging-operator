package exporters

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

type AuthConfig struct {
	Authenticator string `yaml:"authenticator"`
}

// OtlpHttp represents the OpenTelemetry Collector OTLP HTTP exporter configuration
// This is the recommended way to send logs to Loki v3+ using native OTLP ingestion
// See: https://grafana.com/docs/loki/latest/send-data/otel/
// See: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
//
// Example:
//
//	otlphttp := exporters.NewOtlpHttp("", "http://loki:3100/otlp")
//	otlphttp.Encoding = "proto"
//	otlphttp.Compression = "gzip"
//	otlphttp.Headers = map[string]string{
//		"X-Scope-OrgID": "tenant1",
//	}
//	otlphttp.TLS = &types.TlsClientConfig{
//		Insecure: true,
//	}
//	exportersMap := api.Exporters{
//		"otlphttp/loki": otlphttp,
//	}
//	data, err := yaml.Marshal(exportersMap)
//
// produces:
//
//	otlphttp/loki:
//	    endpoint: http://loki:3100/otlp
//	    tls:
//	        insecure: true
//	    headers:
//	        X-Scope-OrgID: tenant1
//	    compression: gzip
//	    encoding: proto
//
// Queue and retry settings can be added via SendingQueue and RetryOnFailure:
//
//	otlphttp.SendingQueue = &types.QueueSettings{
//		Enabled:      true,
//		NumConsumers: 5,
//		QueueSize:    1000,
//	}
//	otlphttp.RetryOnFailure = &types.RetrySettings{
//		Enabled:         true,
//		InitialInterval: "5s",
//		MaxInterval:     "30s",
//		MaxElapsedTime:  "5m",
//	}
//
// which serialize as:
//
//	sending_queue:
//	    enabled: true
//	    num_consumers: 5
//	    queue_size: 1000
//	retry_on_failure:
//	    enabled: true
//	    initial_interval: 5s
//	    max_interval: 30s
//	    max_elapsed_time: 5m
type OtlpHttp struct {
	// id is the exporter identifier in the format "type" or "type/name"
	// It is not serialized to YAML
	id string

	// Endpoint is the target URL for the OTLP HTTP exporter
	// For Loki v3+, use: http://loki:3100/otlp
	Endpoint string `yaml:"endpoint"`

	// TLS configuration
	TLS *types.TlsClientConfig `yaml:"tls,omitempty"`

	// Auth references an authenticator extension
	Auth *AuthConfig `yaml:"auth,omitempty"`

	// Timeout for HTTP requests (default: "30s")
	Timeout string `yaml:"timeout,omitempty"`

	// Headers to include in HTTP requests
	Headers map[string]string `yaml:"headers,omitempty"`

	// Compression for HTTP requests (default: "gzip")
	// Options: "gzip", "none"
	Compression string `yaml:"compression,omitempty"`

	// Encoding specifies the encoding format for OTLP (default: "proto")
	// Options: "proto", "json"
	Encoding string `yaml:"encoding,omitempty"`

	// ReadBufferSize for HTTP client (bytes)
	ReadBufferSize int `yaml:"read_buffer_size,omitempty"`

	// WriteBufferSize for HTTP client (bytes)
	WriteBufferSize int `yaml:"write_buffer_size,omitempty"`

	// MaxIdleConns for HTTP client connection pool
	MaxIdleConns int `yaml:"max_idle_conns,omitempty"`

	// MaxIdleConnsPerHost for HTTP client connection pool
	MaxIdleConnsPerHost int `yaml:"max_idle_conns_per_host,omitempty"`

	// MaxConnsPerHost limits the total number of connections per host
	MaxConnsPerHost int `yaml:"max_conns_per_host,omitempty"`

	// IdleConnTimeout is the maximum amount of time an idle connection will remain idle
	IdleConnTimeout string `yaml:"idle_conn_timeout,omitempty"`

	// Queue configuration for sending data
	SendingQueue *types.QueueSettings `yaml:"sending_queue,omitempty"`

	// Retry configuration for failed requests
	RetryOnFailure *types.RetrySettings `yaml:"retry_on_failure,omitempty"`
}

// ID returns the exporter identifier in the format "type" or "type/name"
func (e *OtlpHttp) ID() string {
	return e.id
}

// ExporterType extracts the exporter type from the ID
func (e *OtlpHttp) ExporterType() types.ExporterType {
	componentType, _ := types.ParseComponentID(e.id)
	return types.ExporterType(componentType)
}

// NewOtlpHttp creates a new OtlpHttp exporter with the given name and endpoint
// If name is empty, the exporter ID will be just the type ("otlphttp")
// If name is provided, the exporter ID will be "otlphttp/name"
// For Loki v3+, use the OTLP endpoint: http://loki:3100/otlp
func NewOtlpHttp(name, endpoint string) *OtlpHttp {
	return &OtlpHttp{
		id:       types.MakeComponentID(string(types.ExporterTypeOtlpHttp), name),
		Endpoint: endpoint,
	}
}
