package input

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func NewViaqReceiverSource(spec *adapters.Input, resNames factory.ForwarderResourceNames, secrets observability.Secrets, op utils.Options) (id string, source types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
	base := helpers.MakeInputID(spec.Name)
	metaID := helpers.MakeID(base, "meta")

	serverTls := tls.NewTlsEnabled(spec, secrets, op)
	switch spec.Receiver.Type {
	case obs.ReceiverTypeSyslog:
		server := sources.NewSyslogServer(helpers.ListenOnAllLocalInterfacesAddress(), spec.Receiver.Port, sources.SyslogModeTcp)
		server.TLS = serverTls
		tfs[metaID] = NewReceiverInternalNormalization(obs.ReceiverTypeSyslog, setEnvelopeToStructured, base)
		spec.Ids = append(spec.Ids, metaID)
		return base, server, tfs
	case obs.ReceiverTypeHTTP:
		itemsID := helpers.MakeID(base, "items")
		tfs[itemsID] = newItemsTransform(base, base)
		server := sources.NewHttpServer(helpers.ListenOnAllLocalInterfacesAddress(), spec.Receiver.Port)
		server.TLS = serverTls
		server.Decoding = &sources.Decoding{
			Codec: codec.CodecTypeJSON,
		}
		tfs[metaID] = NewAuditInternalNormalization(obs.AuditSourceKube, itemsID, false)
		spec.Ids = append(spec.Ids, metaID)
		return base, server, tfs
	default:
		panic(fmt.Sprintf("Unsupported receiver type %q", spec.Receiver.Type))
	}
}

func newItemsTransform(id, inputs string) types.Transform {
	return transforms.NewRemap(`
if exists(.items) {
  r = array([])
  for_each(array!(.items)) -> |_index, i| {
    r = push(r, {"structured": i})
  }
  . = r
} else {
  . = {"structured": .}
}
`, inputs)
}
