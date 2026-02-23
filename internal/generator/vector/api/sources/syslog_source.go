package sources

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
)

type Syslog struct {
	Type    SourceType `json:"type" yaml:"type" toml:"type"`
	Address string     `json:"address" yaml:"address" toml:"address"`
	Mode    SyslogMode `json:"mode" yaml:"mode" toml:"mode"`

	TLS *api.TLS `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func NewSyslogServer(listenAddress string, listenPort int32, mode SyslogMode) *Syslog {
	return &Syslog{
		Type:    SourceTypeSyslog,
		Address: fmt.Sprintf("%s:%d", listenAddress, listenPort),
		Mode:    mode,
	}
}

type SyslogMode string

const (
	SyslogModeTcp SyslogMode = "tcp"
)
