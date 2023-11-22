package input

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vector "github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewViaqReceiverSource(spec logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	base := helpers.MakeInputID(spec.Name)
	var el []generator.Element
	var id string
	switch {
	case logging.IsSyslogReceiver(&spec):
		el = append(el, source.NewSyslogSource(base, spec, op))
		dropID := helpers.MakeID(base, "drop", "debug")
		el = append(el, vector.DropJournalDebugLogs(base, dropID)...)
		id = helpers.MakeID(base, "journal", "viaq")
		el = append(el, vector.JournalLogs(dropID, id)...)
	case logging.IsAuditHttpReceiver(&spec):
		el = []generator.Element{source.NewHttpSource(base, spec, op)}
		id = helpers.MakeID(base, "viaq")
		el = append(el, vector.NormalizeK8sAuditLogs(helpers.MakeID(base, "items"), id)...)
	}
	return el, []string{id}
}
