package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type MethodType string

const (
	MethodTypePost MethodType = "post"
)

type Http struct {
	Type          types.SinkType `json:"type,omitempty"`
	Inputs        []string       `json:"inputs,omitempty"`
	URI           string         `json:"uri,omitempty"`
	Method        MethodType     `json:"method,omitempty"`
	PayloadPrefix string         `json:"payload_prefix,omitempty"`
	PayloadSuffix string         `json:"payload_suffix,omitempty"`
	BaseSink
	Framing *Framing  `json:"framing,omitempty"`
	Auth    *HttpAuth `json:"auth,omitempty"`
	Proxy   *Proxy    `json:"proxy,omitempty"`
}

func NewHttp(url string, init func(s *Http), inputs ...string) (s *Http) {
	sort.Strings(inputs)
	s = &Http{
		Type:   types.SinkTypeHttp,
		Inputs: inputs,
		URI:    url,
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *Http) SinkType() types.SinkType {
	return s.Type
}
