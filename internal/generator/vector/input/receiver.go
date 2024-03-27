package input

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewViaqReceiverSource(spec logging.InputSpec, resNames *factory.ForwarderResourceNames, op generator.Options) ([]generator.Element, []string) {
	base := helpers.MakeInputID(spec.Name)
	var els []generator.Element
	metaID := helpers.MakeID(base, "meta")
	switch {
	case spec.Receiver.IsSyslogReceiver():
		els = append(els,
			source.NewSyslogSource(base, resNames.GenerateInputServiceName(spec.Name), spec, op),
			NewLogSourceAndType(metaID, logging.InfrastructureSourceNode, logging.InputNameInfrastructure, base),
		)
	case spec.Receiver.IsAuditHttpReceiver():
		el, id := source.NewHttpSource(base, resNames.GenerateInputServiceName(spec.Name), spec, op)
		return []generator.Element{
			el,
			NewLogSourceAndType(metaID, logging.AuditSourceKube, logging.InputNameAudit, id),
		}, []string{id}
	}
	return els, []string{metaID}
}
