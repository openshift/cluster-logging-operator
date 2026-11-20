package sinks

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
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
	Enabled bool `json:"enabled,omitempty"`
}

type Batch struct {
	MaxBytes   uint64  `json:"max_bytes,omitempty"`
	MaxEvents  uint    `json:"max_events,omitempty"`
	TimeoutSec float64 `json:"timeout_secs,omitempty"`
}

type Buffer struct {
	Type      BufferType         `json:"type,omitempty"`
	WhenFull  BufferWhenFullType `json:"when_full,omitempty"`
	MaxSize   uint               `json:"max_size,omitempty"`
	MaxEvents uint               `json:"max_events,omitempty"`
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
	Codec           codec.CodecType `json:"codec,omitempty"`
	TimestampFormat string          `json:"timestamp_format,omitempty"`
	ExceptFields    []string        `json:"except_fields,omitempty"`
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
	Method             FramingMethod `json:"method,omitempty"`
	CharacterDelimiter string        `json:"character_delimited,omitempty"`
	MaxLength          uint          `json:"max_length,omitempty"`
}

type Proxy struct {
	Enabled bool   `json:"enabled,omitempty"`
	Http    string `json:"http,omitempty"`
	Https   string `json:"https,omitempty"`
}

type Request struct {
	RetryAttempts           uint              `json:"retry_attempts,omitempty"`
	RetryInitialBackoffSecs uint              `json:"retry_initial_backoff_secs,omitempty"`
	RetryMaxDurationSec     uint              `json:"retry_max_duration_secs,omitempty"`
	TimeoutSecs             uint              `json:"timeout_secs,omitempty"`
	Headers                 map[string]string `json:"headers,omitempty"`
}

type BaseSink struct {
	Compression      CompressionType   `json:"compression,omitempty"`
	Encoding         *Encoding         `json:"encoding,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty"`
	Batch            *Batch            `json:"batch,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty"`
	Request          *Request          `json:"request,omitempty"`
	TLS              *transport.TLS    `json:"tls,omitempty"`
}
