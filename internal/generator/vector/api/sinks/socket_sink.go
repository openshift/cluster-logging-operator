package sinks

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
)

type SocketMode string

const (
	SocketModeTCP          SocketMode = "tcp"
	SocketModeUDP          SocketMode = "udp"
	SocketModeUnixStream   SocketMode = "unix_stream"
	SocketModeUnixDatagram SocketMode = "unix_datagram"
)

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

type Keepalive struct {
	TimeSecs uint `json:"time_secs,omitempty" yaml:"time_secs,omitempty" toml:"time_secs,omitempty"`
}

type SocketEncoding struct {
	Codec        string   `json:"codec,omitempty" yaml:"codec,omitempty" toml:"codec,omitempty"`
	ExceptFields []string `json:"except_fields,omitempty" yaml:"except_fields,omitempty" toml:"except_fields,omitempty"`
	RFC          string   `json:"rfc,omitempty" yaml:"rfc,omitempty" toml:"rfc,omitempty"`
	Facility     string   `json:"facility,omitempty" yaml:"facility,omitempty" toml:"facility,omitempty"`
	Severity     string   `json:"severity,omitempty" yaml:"severity,omitempty" toml:"severity,omitempty"`
	AppName      string   `json:"app_name,omitempty" yaml:"app_name,omitempty" toml:"app_name,omitempty"`
	MsgID        string   `json:"msg_id,omitempty" yaml:"msg_id,omitempty" toml:"msg_id,omitempty"`
	ProcID       string   `json:"proc_id,omitempty" yaml:"proc_id,omitempty" toml:"proc_id,omitempty"`
	Tag          string   `json:"tag,omitempty" yaml:"tag,omitempty" toml:"tag,omitempty"`
	AddLogSource *bool    `json:"add_log_source,omitempty" yaml:"add_log_source,omitempty" toml:"add_log_source,omitempty"`
	PayloadKey   string   `json:"payload_key,omitempty" yaml:"payload_key,omitempty" toml:"payload_key,omitempty"`
}

type Socket struct {
	Type             SinkType          `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs           []string          `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Address          string            `json:"address,omitempty" yaml:"address,omitempty" toml:"address,omitempty"`
	Mode             SocketMode        `json:"mode,omitempty" yaml:"mode,omitempty" toml:"mode,omitempty"`
	Keepalive        *Keepalive        `json:"keepalive,omitempty" yaml:"keepalive,omitempty" toml:"keepalive,omitempty"`
	Path             string            `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	Encoding         *SocketEncoding   `json:"encoding,omitempty" yaml:"encoding,omitempty" toml:"encoding,omitempty"`
	Framing          *Framing          `json:"framing,omitempty" yaml:"framing,omitempty" toml:"framing,omitempty"`
	SendBufferBytes  uint              `json:"send_buffer_bytes,omitempty" yaml:"send_buffer_bytes,omitempty" toml:"send_buffer_bytes,omitempty"`
	HealthCheck      *HealthCheck      `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty" toml:"healthcheck,omitempty"`
	Acknowledgements *Acknowledgements `json:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" toml:"acknowledgements,omitempty"`
	Buffer           *Buffer           `json:"buffer,omitempty" yaml:"buffer,omitempty" toml:"buffer,omitempty"`
	TLS              *api.TLS          `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func NewSocket(mode SocketMode, init func(s *Socket), inputs ...string) (s *Socket) {
	s = &Socket{
		Type:   SinkTypeSocket,
		Inputs: inputs,
		Mode:   mode,
	}
	if init != nil {
		init(s)
	}
	return s
}
