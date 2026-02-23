package sinks

type MethodType string

const (
	MethodTypePost MethodType = "post"
)

type Http struct {
	Type          SinkType   `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs        []string   `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	URI           string     `json:"uri,omitempty" yaml:"uri,omitempty" toml:"uri,omitempty"`
	Method        MethodType `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	Framing       *Framing   `json:"framing,omitempty" yaml:"framing,omitempty" toml:"framing,omitempty"`
	PayloadPrefix string     `json:"payload_prefix,omitempty" yaml:"payload_prefix,omitempty" toml:"payload_prefix,omitempty"`
	PayloadSuffix string     `json:"payload_suffix,omitempty" yaml:"payload_suffix,omitempty" toml:"payload_suffix,omitempty"`
	Auth          *HttpAuth  `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	BaseSink

	Proxy *Proxy `json:"proxy,omitempty" yaml:"proxy,omitempty" toml:"proxy,omitempty"`
}

func NewHttp(url string, init func(s *Http), inputs ...string) (s *Http) {
	s = &Http{
		Type:   SinkTypeHttp,
		Inputs: inputs,
		URI:    url,
	}
	if init != nil {
		init(s)
	}
	return s
}
