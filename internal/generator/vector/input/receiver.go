package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func NewViaqReceiverSource(spec *observability.Input, resNames factory.ForwarderResourceNames, secrets observability.Secrets, op utils.Options) ([]generator.Element, []string) {
	base := helpers.MakeInputID(spec.Name)
	tlsConfig := tls.NewTls(spec, secrets, op)
	if tlsConfig != nil {
		tlsConfig.Enabled = true
	}

	var els []generator.Element
	metaID := helpers.MakeID(base, "meta")

	switch spec.Receiver.Type {
	case obs.ReceiverTypeSyslog:
		els = append(els,
			api.NewConfig(func(c *api.Config) {
				server := sources.NewSyslogServer(helpers.ListenOnAllLocalInterfacesAddress(), spec.Receiver.Port, sources.SyslogModeTcp)
				server.TLS = tlsConfig
				c.Sources[base] = server
			}),
			NewReceiverInternalNormalization(metaID, obs.ReceiverTypeSyslog, setEnvelopeToStructured, base),
		)
	case obs.ReceiverTypeHTTP:
		items, itemsID := newItemsTransform(base, base)
		els = append(els,
			api.NewConfig(func(c *api.Config) {
				server := sources.NewHttpServer(helpers.ListenOnAllLocalInterfacesAddress(), spec.Receiver.Port)
				server.TLS = tlsConfig
				server.Decoding = &sources.Decoding{
					Codec: api.CodecTypeJSON,
				}
				c.Sources[base] = server
			}),
			items,
			NewAuditInternalNormalization(metaID, obs.AuditSourceKube, itemsID, false),
		)
	}
	return els, []string{metaID}
}

func newItemsTransform(id, inputs string) (generator.Element, string) {
	itemsID := helpers.MakeID(id, "items")
	return remap.New(itemsID, `
if exists(.items) {
    r = array([])
    for_each(array!(.items)) -> |_index, i| {
      r = push(r, {"structured": i})
    }
    . = r
} else {
  . = {"structured": .}
}
`, inputs), itemsID
}
