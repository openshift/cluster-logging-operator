package sources

type JournalD struct {
	// Type is required to be 'journald'
	Type SourceType `json:"type" yaml:"type" toml:"type"`

	JournalDirectory string `json:"journal_directory" yaml:"journal_directory" toml:"journal_directory"`
}

func NewJournalD() JournalD {
	return JournalD{
		Type:             SourceTypeJournalD,
		JournalDirectory: "/var/log/journal",
	}
}
