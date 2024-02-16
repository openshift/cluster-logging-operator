package stats

import (
	"sort"
	"time"
)

type Statistics struct {
	Logs      PerfLogs
	MsgSize   int
	TimeDiffs []float64
	Elapsed   time.Duration
	Losses    LossStats
}

func NewStatisics(logs PerfLogs, msgSize int, elapsed time.Duration) *Statistics {
	return &Statistics{
		logs,
		msgSize,
		sortLogsByTimeDiff(logs),
		elapsed,
		NewLossStats(logs),
	}
}
func (stats *Statistics) TotMessages() int {
	return len(stats.Logs)
}

func (stats *Statistics) MeanBloat() float64 {
	return stats.GenericMean((*PerfLog).Bloat)
}

func (stats *Statistics) Mean() float64 {
	return stats.GenericMean((*PerfLog).ElapsedEpoc)
}

func (stats *Statistics) GenericMean(f func(l *PerfLog) float64) float64 {
	if len(stats.Logs) == 0 {
		return 0
	}
	var total float64
	for i := range stats.Logs {
		total += f(&stats.Logs[i])
	}
	return total / float64(len(stats.Logs))
}

func (stats *Statistics) Median() float64 {
	if len(stats.TimeDiffs) == 0 {
		return 0
	}
	return stats.TimeDiffs[(len(stats.TimeDiffs) / 2)]
}

func (stats *Statistics) Min() float64 {
	if len(stats.TimeDiffs) == 0 {
		return 0
	}
	return stats.TimeDiffs[0]
}

func (stats *Statistics) Max() float64 {
	if len(stats.TimeDiffs) == 0 {
		return 0
	}
	return stats.TimeDiffs[len(stats.TimeDiffs)-1]
}

func sortLogsByTimeDiff(logs PerfLogs) []float64 {
	diffs := make([]float64, len(logs))
	for i := range logs {
		diffs[i] = logs[i].ElapsedEpoc()
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i] < diffs[j] })
	return diffs
}
