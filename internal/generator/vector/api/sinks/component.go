package sinks

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
)

type SinkType string

const (
	SinkTypeAwsCloudwatchLogs  SinkType = "aws_cloudwatch_logs"
	SinkTypeAwsS3              SinkType = "aws_s3"
	SinkTypeAzureMonitorLogs   SinkType = "azure_monitor_logs"
	SinkTypeElasticsearch      SinkType = "elasticsearch"
	SinkTypeGcpStackdriverLogs SinkType = "gcp_stackdriver_logs"
	SinkTypeHttp               SinkType = "http"
	SinkTypeLoki               SinkType = "loki"
	SinkTypeKafka              SinkType = "kafka"
	SinkTypeOpenTelemetry      SinkType = "opentelemetry"
	SinkTypePrometheusExporter SinkType = "prometheus_exporter"
	SinkTypeSocket             SinkType = "socket"
	SinkTypeHecLogs            SinkType = "splunk_hec_logs"
)

type CompressionType string

const (
	CompressionTypeGzip   CompressionType = "gzip"
	CompressionTypeNone   CompressionType = "none"
	CompressionTypeSnappy CompressionType = "snappy"
	CompressionTypeZlib   CompressionType = "zlib"
	CompressionTypeZstd   CompressionType = "zstd"
)

type Acknowledgements struct {
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`
}

type Batch struct {
	MaxBytes   uint64  `json:"max_bytes,omitempty" yaml:"max_bytes,omitempty" toml:"max_bytes,omitempty"`
	MaxEvents  uint    `json:"max_events,omitempty" yaml:"max_events,omitempty" toml:"max_events,omitempty"`
	TimeoutSec float64 `json:"timeout_secs,omitempty" yaml:"timeout_secs,omitempty" toml:"timeout_secs,omitempty"`
}

type Buffer struct {
	Type      BufferType         `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	WhenFull  BufferWhenFullType `json:"when_full,omitempty" yaml:"when_full,omitempty" toml:"when_full,omitempty"`
	MaxSize   uint               `json:"max_size,omitempty" yaml:"max_size,omitempty" toml:"max_size,omitempty"`
	MaxEvents uint               `json:"max_events,omitempty" yaml:"max_events,omitempty" toml:"max_events,omitempty"`
}

type BufferType string
type BufferWhenFullType string

const (
	BufferTypeDisk   BufferType = "disk"
	BufferTypeMemory BufferType = "memory"

	BufferWhenFullBlock      BufferWhenFullType = "block"
	BufferWhenFullDropNewest BufferWhenFullType = "drop_newest"
)

type Encoding struct {
	Codec           api.CodecType `json:"codec,omitempty" yaml:"codec,omitempty" toml:"codec,omitempty"`
	TimestampFormat string        `json:"timestamp_format,omitempty" yaml:"timestamp_format,omitempty" toml:"timestamp_format,omitempty"`
	ExceptFields    []string      `json:"except_fields,omitempty" yaml:"except_fields,omitempty" toml:"except_fields,omitempty"`
}

type FramingMethod string

const (
	FramingMethodBytes                 FramingMethod = "bytes"
	FramingMethodCharacterDelimited    FramingMethod = "character_delimited"
	FramingMethodLengthDelimited       FramingMethod = "length_delimited"
	FramingMethodNewlineDelimited      FramingMethod = "newline_delimited"
	FramingMethodVarintLengthDelimited FramingMethod = "varint_length_delimited"
)

type Framing struct {
	Method             FramingMethod `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	CharacterDelimiter string        `json:"character_delimited,omitempty" yaml:"character_delimited,omitempty" toml:"character_delimited,omitempty"`
	MaxLength          uint          `json:"max_length,omitempty" yaml:"max_length,omitempty" toml:"max_length,omitempty"`
}

type Proxy struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" toml:"enabled,omitempty"`
	Http    string `json:"http,omitempty" yaml:"http,omitempty" toml:"http,omitempty"`
	Https   string `json:"https,omitempty" yaml:"https,omitempty" toml:"https,omitempty"`
}

type Request struct {
	RetryAttempts           uint              `json:"retry_attempts,omitempty" yaml:"retry_attempts,omitempty" toml:"retry_attempts,omitempty"`
	RetryInitialBackoffSecs uint              `json:"retry_initial_backoff_secs,omitempty" yaml:"retry_initial_backoff_secs,omitempty" toml:"retry_initial_backoff_secs,omitempty"`
	RetryMaxDurationSec     uint              `json:"retry_max_duration_secs,omitempty" yaml:"retry_max_duration_secs,omitempty" toml:"retry_max_duration_secs,omitempty"`
	TimeoutSecs             uint              `json:"timeout_secs,omitempty" yaml:"timeout_secs,omitempty" toml:"timeout_secs,omitempty"`
	Headers                 map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
}

type BaseSink struct {
	Compression      CompressionType   `json:"compression,omitempty" yaml:"compression,omitempty" toml:"compression,omitempty"`
	Encoding         *Encoding         `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
	TLS              *api.TLS          `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}
