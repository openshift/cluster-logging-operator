package input

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewViaqJournalSource(input logging.InputSpec) ([]Element, []string) {
	id := helpers.MakeInputID(input.Name, "journal")
	el := []Element{
		source.NewJournalLog(id),
	}
	dropID := helpers.MakeID(id, "drop")
	el = append(el, viaq.DropJournalDebugLogs(id, dropID)...)
	normalizeID := helpers.MakeID(id, "viaq")
	el = append(el, viaq.JournalLogs(dropID, normalizeID)...)
	return el, []string{normalizeID}
}
