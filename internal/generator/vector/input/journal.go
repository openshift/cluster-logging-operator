package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	v1 "github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type JournalLog struct {
	api.Config
}

func newJournalLog(id string) JournalLog {
	return JournalLog{
		Config: api.Config{
			Sources: map[string]interface{}{
				id: sources.NewJournalD(),
			},
		},
	}
}

func NewJournalInput(input obs.InputSpec) ([]framework.Element, []string) {
	id := helpers.MakeInputID(input.Name, "journal")
	metaID := helpers.MakeID(id, "meta")
	el := []framework.Element{
		newJournalLog(id),
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
