package main

import (
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"sort"
	"time"
)

type Statistics struct {
	logs      types.PerfLogs
	msgSize   int
	timeDiffs []float64
	elapsed   float64
}

func NewStatisics(logs types.PerfLogs, msgSize int, elapsed time.Duration) *Statistics {
	return &Statistics{
		logs,
		msgSize,
		sortLogsByTimeDiff(logs),
		elapsed.Seconds(),
	}
}
func (stats *Statistics) TotMessages() int {
	return len(stats.logs)
}

func (stats *Statistics) meanBloat() float64 {
	return stats.genericMean((*types.PerfLog).Bloat)
}

func (stats *Statistics) mean() float64 {
	return stats.genericMean((*types.PerfLog).ElapsedEpoc)
}

func (stats *Statistics) genericMean(f func(l *types.PerfLog) float64) float64 {
	if len(stats.logs) == 0 {
		return 0
	}
	var total float64
	for i := range stats.logs {
		total += f(&stats.logs[i])
	}
	return total / float64(len(stats.logs))
}

func (stats *Statistics) median() float64 {
	if len(stats.timeDiffs) == 0 {
		return 0
	}
	return stats.timeDiffs[(len(stats.timeDiffs) / 2)]
}

func (stats *Statistics) min() float64 {
	if len(stats.timeDiffs) == 0 {
		return 0
	}
	return stats.timeDiffs[0]
}

func (stats *Statistics) max() float64 {
	if len(stats.timeDiffs) == 0 {
		return 0
	}
	return stats.timeDiffs[len(stats.timeDiffs)-1]
}

func sortLogsByTimeDiff(logs types.PerfLogs) []float64 {
	diffs := make([]float64, len(logs))
	for i := range logs {
		diffs[i] = logs[i].ElapsedEpoc()
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i] < diffs[j] })
	return diffs
}
