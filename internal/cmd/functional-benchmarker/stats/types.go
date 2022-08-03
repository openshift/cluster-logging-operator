package stats

import (
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

type PerfLog struct {
	types.AllLog
	types.Timing `json:",inline"`
	SequenceId   int
}

type PerfLogs []PerfLog

func (t *PerfLog) ElapsedEpoc() float64 {
	return t.EpocOut - t.EpocIn
}
