package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewJournalSource(input obs.InputSpec) ([]Element, []string) {
	id := helpers.MakeInputID(input.Name, "journal")
	metaID := helpers.MakeID(id, "meta")
	el := []Element{
		source.NewJournalLog(id),
		NewInternalNormalization(metaID, string(obs.InfrastructureSourceNode), string(obs.InputTypeInfrastructure), id,
			viaq.SystemK,
			viaq.SystemT,
			viaq.SystemU,
		),
	}
	return el, []string{metaID}
}
