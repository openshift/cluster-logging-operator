package input

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vector "github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewViaqReceiverSource(spec logging.InputSpec, resNames *factory.ForwarderResourceNames, op generator.Options) ([]generator.Element, []string) {
	base := helpers.MakeInputID(spec.Name)
	var el []generator.Element
	var id string
	switch {
	case spec.Receiver.IsSyslogReceiver():
		el = append(el, source.NewSyslogSource(base, resNames.GenerateInputServiceName(spec.Name), spec, op))
		dropID := helpers.MakeID(base, "drop", "debug")
		el = append(el, vector.DropJournalDebugLogs(base, dropID)...)
		id = helpers.MakeID(base, "journal", "viaq")
		el = append(el, vector.JournalLogs(dropID, id)...)
	case spec.Receiver.IsAuditHttpReceiver():
		el = []generator.Element{source.NewHttpSource(base, resNames.GenerateInputServiceName(spec.Name), spec, op)}
		id = helpers.MakeID(base, "viaq")
		el = append(el, vector.NormalizeK8sAuditLogs(helpers.MakeID(base, "items"), id)...)
	}
	return el, []string{id}
}
