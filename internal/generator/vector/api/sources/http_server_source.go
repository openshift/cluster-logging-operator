package sources

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
)

type HttpServer struct {
	Type     types.SourceType `json:"type" yaml:"type" toml:"type"`
	Address  string           `json:"address" yaml:"address" toml:"address"`
	Decoding *Decoding        `json:"decoding,omitempty" yaml:"decoding,omitempty" toml:"decoding,omitempty"`

	TLS *transport.TlsEnabled `json:"tls,omitempty" yaml:"tls,omitempty" toml:"tls,omitempty"`
}

func (h HttpServer) SourceType() types.SourceType {
	return h.Type
}

func NewHttpServer(listenAddress string, listenPort int32) *HttpServer {
	return &HttpServer{
		Type:    types.SourceTypeHttpServer,
		Address: fmt.Sprintf("%s:%d", listenAddress, listenPort),
	}
}

type Decoding struct {
	Codec codec.CodecType `json:"codec,omitempty" yaml:"codec,omitempty" toml:"codec,omitempty"`
}
