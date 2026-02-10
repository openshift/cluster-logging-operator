package sources

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
)

type HttpServer struct {
	Type     SourceType `json:"type" yaml:"type" toml:"type"`
	Address  string     `json:"address" yaml:"address" toml:"address"`
	Decoding *Decoding  `json:"decoding,omitempty" yaml:"decoding,omitempty" toml:"decoding,omitempty"`

	TLS *api.TLS `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func NewHttpServer(listenAddress string, listenPort int32) *HttpServer {
	return &HttpServer{
		Type:    SourceTypeHttpServer,
		Address: fmt.Sprintf("%s:%d", listenAddress, listenPort),
	}
}

type Decoding struct {
	Codec api.CodecType `json:"codec,omitempty" yaml:"codec,omitempty" toml:"codec,omitempty"`
}
