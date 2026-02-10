package sinks

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type SinkType string

const (
	SinkTypeOpenTelemetry SinkType = "opentelemetry"
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

func NewBatch(maxBytes *resource.Quantity) (b *Batch) {
	if maxBytes != nil && !maxBytes.IsZero() {
		b = &Batch{
			MaxBytes: uint64(maxBytes.Value()),
		}
	}
	return b
}

type Buffer struct {
	Type      BufferType         `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	WhenFull  BufferWhenFullType `json:"when_full,omitempty" yaml:"when_full,omitempty" toml:"when_full,omitempty"`
	MaxSize   uint               `json:"max_size,omitempty" yaml:"max_size,omitempty" toml:"max_size,omitempty"`
	MaxEvents uint               `json:"max_events,omitempty" yaml:"max_events,omitempty" toml:"max_events,omitempty"`
}

func NewBuffer(init func(buffer *Buffer)) (b *Buffer) {
	b = &Buffer{}
	if init != nil {
		init(b)
	}
	return b
}

type BufferType string
type BufferWhenFullType string

const (
	BufferTypeDisk   BufferType = "disk"
	BufferTypeMemory BufferType = "memory"

	BufferWhenFullBlock      BufferWhenFullType = "block"
	BufferWhenFullDropNewest BufferWhenFullType = "drop_newest"
)

type Request struct {
	RetryAttempts           uint `json:"retry_attempts,omitempty" yaml:"retry_attempts,omitempty" toml:"retry_attempts,omitempty"`
	RetryInitialBackoffSecs uint `json:"retry_initial_backoff_secs,omitempty" yaml:"retry_initial_backoff_secs,omitempty" toml:"retry_initial_backoff_secs,omitempty"`
	RetryMaxDurationSec     uint `json:"retry_max_duration_secs,omitempty" yaml:"retry_max_duration_secs,omitempty" toml:"retry_max_duration_secs,omitempty"`
}

func NewRequest(init func(r *Request)) *Request {
	r := &Request{}
	if init != nil {
		init(r)
	}
	return r
}

type BaseSink struct {
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty" yaml:"batch,omitempty" toml:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty" yaml:"request,omitempty" toml:"request,omitempty"`
}
