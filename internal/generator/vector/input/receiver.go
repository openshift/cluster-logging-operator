package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func NewViaqReceiverSource(spec obs.InputSpec, resNames factory.ForwarderResourceNames, secrets observability.Secrets, op generator.Options) ([]generator.Element, []string) {
	base := helpers.MakeInputID(spec.Name)
	tlsConfig := newInputTLS(spec.Receiver.TLS, secrets, op)

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
					Codec: sources.CodecTypeJSON,
				}
				c.Sources[base] = server
			}),
			items,
			NewAuditInternalNormalization(metaID, obs.AuditSourceKube, itemsID, false),
		)
	}
	return els, []string{metaID}
}

func newInputTLS(spec *obs.InputTLSSpec, secrets observability.Secrets, op utils.Options) *api.TLS {
	if spec == nil {
		return nil
	}
	inputTls := &api.TLS{
		Enabled: true,
		KeyFile: tls.SecretPath(spec.Key, "%s"),
		CRTFile: tls.ValuePath(spec.Certificate, "%s"),
		CAFile:  tls.ValuePath(spec.CA, "%s"),
		KeyPass: secrets.AsString(spec.KeyPassphrase),
	}
	inputTls.SetTLSProfile(op)
	return inputTls
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
