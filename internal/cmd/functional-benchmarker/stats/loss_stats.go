package stats

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
)

type LossStats struct {
	raw             PerfLogs
	entriesByLoader map[string][]PerfLog
}

type StreamLossStats struct {
	MinSeqId  int
	MaxSeqId  int
	Collected int
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

// Lost is the number of messages not collected for the stream
func (s *StreamLossStats) Lost() int {
	return s.MaxSeqId - s.MinSeqId - s.Collected
}

// Range is difference between the first and last collected sequence ids
func (s *StreamLossStats) Range() int {
	return s.MaxSeqId - s.MinSeqId
}

func (l *LossStats) LossStatsFor(stream string) (*StreamLossStats, error) {
	l.init()
	entries, found := l.entriesByLoader[stream]
	if !found {
		return nil, fmt.Errorf("No lost entries found for %s", stream)
	}
	sort.Slice(entries, func(l int, r int) bool {
		left, err := GetSequenceIdFrom(entries[l].Message)
		if err != nil {
			log.Error(err, "message", entries[l].Message)
			return false
		}
		right, err := GetSequenceIdFrom(entries[r].Message)
		if err != nil {
			log.Error(err, "message", entries[r].Message)
			return false
		}
		return left < right
	})

	log.V(4).Info("Retrieving the seqId from first message received", "message", entries[0].Message)
	min, errMin := GetSequenceIdFrom(entries[0].Message)
	log.V(4).Info("Retrieving the seqId from last message evaluated", "messag", entries[len(entries)-1].Message)
	max, errMax := GetSequenceIdFrom(entries[len(entries)-1].Message)

	missing := 0
	next := min
	for _, entry := range entries {
		seq, err := GetSequenceIdFrom(entry.Message)
		if err != nil {
			log.Error(err, "Error evaluating seqId", "message", entry.Message)
		} else {
			if next != seq {
				missing += (seq - next)
			}
			next = seq + 1
		}
	}

	var err error
	if errMin != nil || errMax != nil {
		err = fmt.Errorf("minError: %v, maxError: %v", errMin, errMax)
	}
	lossStats := StreamLossStats{
		MinSeqId:  min,
		MaxSeqId:  max,
		Collected: max - min - missing,
		Entries:   entries,
	}
	return &lossStats, err
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
		streams, found := results[entry.Kubernetes.ContainerName]
		if !found {
			streams = []PerfLog{}
		}
		streams = append(streams, entry)
		results[entry.Kubernetes.ContainerName] = streams
	}

	return results
}

func GetSequenceIdFrom(message string) (int, error) {

	parts := strings.Split(message, " - ")
	log.V(4).Info("Evaluating seq from parts", "parts", parts)
	if len(parts) >= 3 {
		seq := parts[2]
		return strconv.Atoi(seq)
	}
	return 0, fmt.Errorf("message is not the expected format containing sequence number: %s", message)
}
