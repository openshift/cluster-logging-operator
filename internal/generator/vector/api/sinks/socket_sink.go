package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type SocketMode string

const (
	SocketModeTCP          SocketMode = "tcp"
	SocketModeUDP          SocketMode = "udp"
	SocketModeUnixStream   SocketMode = "unix_stream"
	SocketModeUnixDatagram SocketMode = "unix_datagram"
)

type Keepalive struct {
	TimeSecs uint `json:"time_secs,omitempty"`
}

type SyslogEncodingConfig struct {
	RFC      string `json:"rfc,omitempty"`
	Facility string `json:"facility,omitempty"`
	Severity string `json:"severity,omitempty"`
	AppName  string `json:"app_name,omitempty"`
	MsgID    string `json:"msg_id,omitempty"`
	ProcID   string `json:"proc_id,omitempty"`
}

type SocketEncoding struct {
	Codec        string                `json:"codec,omitempty"`
	ExceptFields []string              `json:"except_fields,omitempty"`
	Syslog       *SyslogEncodingConfig `json:"syslog,omitempty"`
}

type Socket struct {
	Type             types.SinkType        `json:"type,omitempty"`
	Inputs           []string              `json:"inputs,omitempty"`
	Address          string                `json:"address,omitempty"`
	Mode             SocketMode            `json:"mode,omitempty"`
	Keepalive        *Keepalive            `json:"keepalive,omitempty"`
	Path             string                `json:"path,omitempty"`
	Encoding         *SocketEncoding       `json:"encoding,omitempty"`
	Framing          *Framing              `json:"framing,omitempty"`
	SendBufferBytes  uint                  `json:"send_buffer_bytes,omitempty"`
	HealthCheck      *HealthCheck          `json:"healthcheck,omitempty"`
	Acknowledgements *Acknowledgements     `json:"acknowledgements,omitempty"`
	Buffer           *Buffer               `json:"buffer,omitempty"`
	TLS              *transport.TlsEnabled `json:"tls,omitempty"`
}

func NewSocket(mode SocketMode, init func(s *Socket), inputs ...string) (s *Socket) {
	sort.Strings(inputs)
	s = &Socket{
		Type:   types.SinkTypeSocket,
		Inputs: inputs,
		Mode:   mode,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *Socket) SinkType() types.SinkType {
	return s.Type
}
