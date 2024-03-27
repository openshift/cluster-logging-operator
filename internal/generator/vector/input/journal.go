package input

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewJournalSource(input logging.InputSpec) ([]Element, []string) {
	id := helpers.MakeInputID(input.Name, "journal")
	metaID := helpers.MakeID(id, "meta")
	el := []Element{
		source.NewJournalLog(id),
		NewLogSourceAndType(metaID, logging.InfrastructureSourceNode, logging.InputNameInfrastructure, id),
	}
	return el, []string{metaID}
}
