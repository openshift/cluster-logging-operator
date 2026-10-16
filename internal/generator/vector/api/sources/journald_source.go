package sources

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Journald struct {
	// Type is required to be 'journald'
	Type types.SourceType `json:"type" yaml:"type" toml:"type"`

	JournalDirectory string `json:"journal_directory" yaml:"journal_directory" toml:"journal_directory"`

	// JournalctlPath is pinned to the absolute path of the trusted journalctl binary so the
	// collector never falls back to a PATH search or a config-controlled override, which could
	// otherwise be used to execute an attacker-placed binary (LOG-9759 / FIND-003).
	JournalctlPath string `json:"journalctl_path" yaml:"journalctl_path" toml:"journalctl_path"`
}

func NewJournalD() Journald {
	return Journald{
		Type:             types.SourceTypeJournald,
		JournalDirectory: "/var/log/journal",
		JournalctlPath:   "/usr/bin/journalctl",
	}
}

func (s Journald) SourceType() types.SourceType {
	return s.Type
}
