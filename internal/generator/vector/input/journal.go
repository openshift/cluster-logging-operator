package input

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewViaqJournalSource(input logging.InputSpec) ([]Element, []string) {
	id := helpers.MakeInputID(input.Name, "journal")
	el := []Element{
		source.NewJournalLog(id),
	}
	dropID := helpers.MakeID(id, "drop")
	el = append(el, normalize.DropJournalDebugLogs(id, dropID)...)
	normalizeID := helpers.MakeID(id, "viaq")
	el = append(el, normalize.JournalLogs(dropID, normalizeID)...)
	return el, []string{normalizeID}
}
