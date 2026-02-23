package http

import (
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func New(id string, o *observability.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(helpers.MakeID(id, "debug"), helpers.MakeInputs(inputs...)),
		}
	}
	var els []framework.Element
	return framework.MergeElements(

		els,
		[]framework.Element{
			api.NewConfig(func(c *api.Config) {
				c.Sinks[id] = sinks.NewHttp(o.HTTP.URL, func(s *sinks.Http) {
					s.URI = o.HTTP.URL
					s.Framing = framing(o.HTTP)
					s.Auth = common.NewHttpAuth(o.HTTP.Authentication, op)
					s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
					s.Compression = sinks.CompressionType(o.GetTuning().Compression)
					s.Batch = common.NewApiBatch(o)
					s.Buffer = common.NewApiBuffer(o)
					request(s, o)
					s.Method = method(o.HTTP)
					s.TLS = tls.NewTls(o, secrets, op)
					if o.HTTP.ProxyURL != "" {
						s.Proxy = &sinks.Proxy{
							Enabled: true,
							Http:    o.HTTP.ProxyURL,
							Https:   o.HTTP.ProxyURL,
						}
					}
				}, inputs...)
			}),
		},
	)
}

func framing(h *obs.HTTP) *sinks.Framing {
	if h.Format == obs.HTTPFormatNDJSON {
		return &sinks.Framing{
			Method: sinks.FramingMethodNewlineDelimited,
		}
	}
	return nil
}

func method(h *obs.HTTP) sinks.MethodType {
	if h == nil {
		return sinks.MethodTypePost
	}
	if h.Method == "" {
		return sinks.MethodTypePost
	}
	return sinks.MethodType(strings.ToLower(h.Method))
}

func request(s *sinks.Http, o *observability.Output) {
	s.Request = common.NewApiRequest(o)
	if o.HTTP != nil && o.HTTP.Timeout != 0 {
		if s.Request == nil {
			s.Request = &sinks.Request{}
		}
		s.Request.TimeoutSecs = uint(o.HTTP.Timeout)
	}
	if o.HTTP != nil && len(o.HTTP.Headers) != 0 {
		if s.Request == nil {
			s.Request = &sinks.Request{}
		}
		s.Request.Headers = o.HTTP.Headers
	}
}
