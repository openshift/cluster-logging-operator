package sources

type SourceType string

const (
	SourceTypeFile            SourceType = "file"
	SourceTypeInternalMetrics SourceType = "internal_metrics"
	SourceTypeJournalD        SourceType = "journald"
)
