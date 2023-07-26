package stats

import (
	"fmt"
	"sort"
)

type LossStats struct {
	raw             PerfLogs
	entriesByLoader map[string][]PerfLog
}

type StreamLossStats struct {
	MinSeqId  int
	MaxSeqId  int
	Collected int
	Purged    int
	Entries   []PerfLog
}

func NewLossStats(logs PerfLogs) LossStats {
	return LossStats{
		raw: logs,
	}
}

func (l *LossStats) init() {
	if l.entriesByLoader == nil {
		l.entriesByLoader = splitEntriesByLoader(l.raw)
	}
}

// Range is difference between the first and last collected sequence ids
func (s *StreamLossStats) Range() int {
	return s.MaxSeqId - s.MinSeqId
}

func (l *StreamLossStats) PercentCollected() float64 {
	return float64(l.Collected) / float64(l.Range()) * 100.0
}

func (l *LossStats) LossStatsFor(stream string) (*StreamLossStats, error) {
	l.init()
	entries, found := l.entriesByLoader[stream]
	if !found {
		return nil, fmt.Errorf("No lost entries found for %s", stream)
	}

	sort.Slice(entries, func(l int, r int) bool {
		return entries[l].SequenceId < entries[r].SequenceId
	})

	lossStats := StreamLossStats{
		Collected: len(entries),
		Entries:   entries,
	}
	if len(entries) == 0 {
		return &lossStats, nil
	}

	lossStats.MinSeqId = entries[0].SequenceId
	lossStats.MaxSeqId = entries[len(entries)-1].SequenceId

	return &lossStats, nil
}

func (l *LossStats) Streams() []string {
	l.init()
	streams := []string{}
	for s := range l.entriesByLoader {
		streams = append(streams, s)
	}
	sort.Strings(streams)
	return streams
}

func splitEntriesByLoader(logs PerfLogs) map[string][]PerfLog {
	results := map[string][]PerfLog{}
	for _, entry := range logs {
		streamName := entry.Stream
		streams, found := results[streamName]
		if !found {
			streams = []PerfLog{}
		}
		streams = append(streams, entry)
		results[streamName] = streams
	}

	return results
}
