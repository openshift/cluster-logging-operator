package sources

type SourceType string

const (
	SourceTypeFile            SourceType = "file"
	SourceTypeHttpServer      SourceType = "http_server"
	SourceTypeInternalMetrics SourceType = "internal_metrics"
	SourceTypeKubernetesLogs  SourceType = "kubernetes_logs"
	SourceTypeJournalD        SourceType = "journald"
	SourceTypeSyslog          SourceType = "syslog"
)
