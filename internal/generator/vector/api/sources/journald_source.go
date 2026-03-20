package sources

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Journald struct {
	// Type is required to be 'journald'
	Type types.SourceType `json:"type" yaml:"type" toml:"type"`

	JournalDirectory string `json:"journal_directory" yaml:"journal_directory" toml:"journal_directory"`
}

func NewJournalD() Journald {
	return Journald{
		Type:             types.SourceTypeJournald,
		JournalDirectory: "/var/log/journal",
	}
}

func (s Journald) SourceType() types.SourceType {
	return s.Type
}
