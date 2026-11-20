package sources

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type Syslog struct {
	Type    types.SourceType `json:"type"`
	Address string           `json:"address"`
	Mode    SyslogMode       `json:"mode"`

	TLS *transport.TlsEnabled `json:"tls,omitempty"`
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
