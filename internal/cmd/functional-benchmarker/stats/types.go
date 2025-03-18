package stats

import (
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

type PerfLog struct {
	types.Timing

	// ContainerName is the name of the container producing the logstream
	ContainerName string `json:"container_name,omitempty"`

	// Message is the receiver processed messsage, not the original message
	Message string `json:"message,omitempty"`

	// MessageSize is the original message size before it was processed by the receiver
	MessageSize int `json:"message_size,omitempty"`

	// PayloadSize is the original payload size before it was processed by the receiver
	PayloadSize int `json:"payload_size,omitempty"`

	Stream     string
	SequenceId int `json:"seqid,omitempty"`
}

type PerfLogs []PerfLog

func (t *PerfLog) Bloat() float64 {
	return float64(t.PayloadSize) / float64(t.MessageSize)
}
func (t *PerfLog) ElapsedEpoc() float64 {
	return t.EpocOut - t.EpocIn
}

// NewPerfLog creates a PerfLog from a line parsing it or returning nil if there is an error
func NewPerfLog(line string) *PerfLog {
	entry := &PerfLog{}
	log.V(4).Info("Unmarshalling", "line", line)
	if err := test.Unmarshal(line, entry); err != nil {
		log.V(4).Info("Failed to unmarshall perflog", "line", line)
		return nil
	}
	entry.Stream = entry.ContainerName
	log.V(4).Info("Returning", "perfLog", entry)
	return entry
}
