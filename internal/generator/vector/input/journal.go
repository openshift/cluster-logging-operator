package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	v1 "github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewJournalInput(input *adapters.Input) (id string, source types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
	source = sources.NewJournalD()
	id = helpers.MakeInputID(input.Name, "journal")
	metaID := helpers.MakeID(id, "meta")
	tfs.Add(metaID, NewJournalInternalNormalization(obs.InfrastructureSourceNode, setEnvelope, id,
		v1.FixJournalLogLevel,
		v1.SetJournalMessage,
		v1.SystemK,
		v1.SystemT,
		v1.SystemU,
	))
	input.Ids = append(input.Ids, metaID)
	return id, source, tfs
}
