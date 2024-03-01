package tuning

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	reSeqID = regexp.MustCompile(`\[ [0-9]* \]:`)
)

type StreamStats struct {
	Name        string
	SequenceIDs []int
	Missing     []int
	Duplicates  []int
	First       int
	Last        int
}

type LogStreams map[string]*StreamStats

func (ls LogStreams) Len() int {
	return len(ls)
}
func (ls LogStreams) Add(l types.AllLog) error {
	s, found := ls[l.StreamName()]
	if !found {
		s = &StreamStats{
			Name: l.StreamName(),
		}
		ls[s.Name] = s
	}
	match := reSeqID.FindString(l.Message)
	match = strings.Trim(match, "[]: ")
	seq, err := strconv.Atoi(match)
	if err != nil {
		log.V(0).Error(err, "Unable to extract sequenceID", "message", l.Message)
		return nil
	}
	s.SequenceIDs = append(s.SequenceIDs, seq)
	return nil
}

func (ls LogStreams) Evaluate() {
	log.V(1).Info("Evaluating log streams", "total", len(ls))
	for name, s := range ls {
		if len(s.SequenceIDs) > 0 {
			sort.Ints(s.SequenceIDs)
			log.V(3).Info("Messages received", "name", name, "tot", len(s.SequenceIDs), "minID", s.SequenceIDs[0], "maxID", s.SequenceIDs[len(s.SequenceIDs)-1])
			s.First = s.SequenceIDs[0]
			s.Last = s.SequenceIDs[len(s.SequenceIDs)-1]
			var prev int
			for i, seq := range s.SequenceIDs {
				log.V(4).Info("Evaluating sequence", "stream", name, "seqID", seq, "prev", prev)
				switch {
				case i == 0: //start eval from first message, regardless of sequence
					prev = seq
				case prev == seq: //must be duplicate
					log.V(4).Info("Duplicate Sequence", "stream", name, "seqID", seq, "prev", prev)
					s.Duplicates = append(s.Duplicates, seq)
				case prev+1 != seq: //gap between last and next gathered
					log.V(4).Info("Missing Sequence", "stream", name, "seqID", seq, "prev", prev, "range", seq-prev)
					for i = 1; i <= seq-prev; i++ {
						s.Missing = append(s.Missing, prev+i)
					}
					prev = seq
				default: // next in line
					prev = seq
				}
			}
		} else {
			log.V(0).Info("No messages to evaluate sequences", "stream", s.Name)
		}
	}
}
