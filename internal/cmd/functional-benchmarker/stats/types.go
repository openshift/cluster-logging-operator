package stats

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"regexp"
	"strconv"
	"strings"
)

var (
	perfLogPattern = regexp.MustCompile(`.*epoc_in\":(?P<epoc_in>[0-9\.]*).*epoc_out\":(?P<epoc_out>[0-9\.]*).*message\":\"(?P<message>.*(?P<stream>functional\.0\.[0-9A-Z]*) - (?P<seqid>\d{10}) -.*?\").*?`)
)

type PerfLog struct {
	Stream string
	types.Timing
	SequenceId int

	// Bloat is the ratio of overall size / Message size
	bloat float64
}

type PerfLogs []PerfLog

func (t *PerfLog) Bloat() float64 {
	return t.bloat
}
func (t *PerfLog) ElapsedEpoc() float64 {
	return t.EpocOut - t.EpocIn
}

// NewPerfLog creates a PerfLog from a line parsing it or returning nil if there is an error
func NewPerfLog(line string) *PerfLog {
	match := perfLogPattern.FindStringSubmatch(line)
	if len(match) > 0 {
		entry := PerfLog{}
		for i, name := range perfLogPattern.SubexpNames() {
			switch name {
			case "message":
				entry.bloat = float64(len(line)) / float64(len(match[i]))
			case "stream":
				entry.Stream = match[i]
			case "seqid":
				seqId, err := strconv.Atoi(strings.TrimSpace(match[i]))
				if err != nil {
					log.Error(err, "Skipping entry. Unable to parse sequence id", "match", match[i], "line", line)
					return nil
				}
				entry.SequenceId = seqId
			case "epoc_out":
				epocOut, err := strconv.ParseFloat(strings.TrimSpace(match[i]), 64)
				if err != nil {
					log.Error(err, "Skipping entry. Unable to parse epic_out", "match", match[i], "line", line)
					return nil
				}
				entry.EpocOut = epocOut
			case "epoc_in":
				epicIn, err := strconv.ParseFloat(strings.TrimSpace(match[i]), 64)
				if err != nil {
					log.Error(err, "Skipping entry. Unable to parse epic_in", "match", match[i], "line", line)
					return nil
				}
				entry.EpocIn = epicIn
			}
		}

		return &entry
	}
	log.V(4).Info("Failed to match perflog", "line", line)
	return nil
}
