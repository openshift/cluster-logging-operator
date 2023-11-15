package stats

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"path"
	"strconv"
	"strings"
)

type PerfLog struct {
	Stream string `json:"stream,omitempty"`
	types.Timing
	SequenceId    int
	SequenceIdStr string `json:"seqid,omitempty"`

	// Bloat is the ratio of overall size / Message size
	Bloat float64 `json:"Bloat,omitempty"`
}

type PerfLogs []PerfLog

func (t *PerfLog) ElapsedEpoc() float64 {
	return t.EpocOut - t.EpocIn
}

// NewPerfLog creates a PerfLog from a line parsing it or returning nil if there is an error
func NewPerfLog(line, fileName string) *PerfLog {
	entry := &PerfLog{}
	if err := types.ParseLogsFrom(line, entry, false); err != nil {
		log.Error(err, "Error parsing line into PerfLog", "line", line)
		return nil
	}
	seqId, err := strconv.Atoi(entry.SequenceIdStr)
	if err != nil {
		log.Error(err, "Skipping entry. Unable to parse sequence id", "seqId", entry.SequenceIdStr)
		return nil
	}
	entry.SequenceId = seqId
	entry.Stream = strings.TrimSuffix(path.Base(fileName), ".log")
	return entry
}
