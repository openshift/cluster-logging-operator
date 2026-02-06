package sources

type SourceType string

const (
	SourceTypeFile     SourceType = "file"
	SourceTypeJournalD SourceType = "journald"
)
