package sources

type SourceType string

const (
	SourceTypeFile            SourceType = "file"
	SourceTypeInternalMetrics SourceType = "internal_metrics"
	SourceTypeKubernetesLogs  SourceType = "kubernetes_logs"
	SourceTypeJournalD        SourceType = "journald"
)
