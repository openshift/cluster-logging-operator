package types

type SourceType string

const (
	SourceTypeFile            SourceType = "file"
	SourceTypeHttpServer      SourceType = "http_server"
	SourceTypeInternalMetrics SourceType = "internal_metrics"
	SourceTypeKubernetesLogs  SourceType = "kubernetes_logs"
	SourceTypeJournald        SourceType = "journald"
	SourceTypeSyslog          SourceType = "syslog"
)

// Source is a vector source for signals coming into the collector
type Source interface {
	SourceType() SourceType
}
