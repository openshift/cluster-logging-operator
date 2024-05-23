package input

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewViaqReceiverSource(spec obs.InputSpec, resNames factory.ForwarderResourceNames, secrets helpers.Secrets, op generator.Options) ([]generator.Element, []string) {
	base := helpers.MakeInputID(spec.Name)
	tlsConfig := receiverTLS(base, spec.Receiver.TLS, secrets, op)

	var els []generator.Element
	metaID := helpers.MakeID(base, "meta")

	switch spec.Receiver.Type {
	case obs.ReceiverTypeSyslog:
		els = append(els,
			source.NewSyslogSource(base, resNames.GenerateInputServiceName(spec.Name), spec),
			tlsConfig,
			NewLogSourceAndType(metaID, logging.InfrastructureSourceNode, logging.InputNameInfrastructure, base),
		)
	case obs.ReceiverTypeHTTP:
		el, id := source.NewHttpSource(base, resNames.GenerateInputServiceName(spec.Name), spec)
		split, splitID := source.NewSplitTransform(base, id)
		items, itemsID := source.NewItemsTransform(base, splitID)
		els = append(els,
			el,
			tlsConfig,
			split,
			items,
			NewLogSourceAndType(metaID, logging.AuditSourceKube, logging.InputNameAudit, itemsID),
		)
	}
	return els, []string{metaID}
}

func receiverTLS(id string, spec *obs.InputTLSSpec, secrets helpers.Secrets, op generator.Options) generator.Element {
	if spec == nil {
		return generator.Nil
	}
	tlsSpec := &obs.OutputTLSSpec{
		TLSSpec: obs.TLSSpec{
			CA:            spec.CA,
			Certificate:   spec.Certificate,
			Key:           spec.Key,
			KeyPassphrase: spec.KeyPassphrase,
		},
	}
	template := tls.New(id, tlsSpec, secrets, op)
	template.Component = "sources"
	return template
}
