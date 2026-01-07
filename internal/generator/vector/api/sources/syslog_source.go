package sources

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type Syslog struct {
	Type    types.SourceType `json:"type" yaml:"type" toml:"type"`
	Address string           `json:"address" yaml:"address" toml:"address"`
	Mode    SyslogMode       `json:"mode" yaml:"mode" toml:"mode"`

	TLS *transport.TlsEnabled `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func (s Syslog) SourceType() types.SourceType {
	return s.Type
}

func NewSyslogServer(listenAddress string, listenPort int32, mode SyslogMode) *Syslog {
	return &Syslog{
		Type:    types.SourceTypeSyslog,
		Address: fmt.Sprintf("%s:%d", listenAddress, listenPort),
		Mode:    mode,
	}
}

type SyslogMode string

const (
	SyslogModeTcp SyslogMode = "tcp"
)
