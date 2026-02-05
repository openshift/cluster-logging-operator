package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewJournalSource(input obs.InputSpec) ([]framework.Element, []string) {
	id := helpers.MakeInputID(input.Name, "journal")
	metaID := helpers.MakeID(id, "meta")
	el := []framework.Element{
		source.NewJournalLog(id),
		NewJournalInternalNormalization(metaID, obs.InfrastructureSourceNode, setEnvelope, id,
			v1.FixJournalLogLevel,
			v1.SetJournalMessage,
			v1.SystemK,
			v1.SystemT,
			v1.SystemU,
		),
	}
	return el, []string{metaID}
}
