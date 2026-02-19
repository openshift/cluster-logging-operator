package types

type SinkType string

const (
	SinkTypeAwsCloudwatchLogs  SinkType = "aws_cloudwatch_logs"
	SinkTypeAwsS3              SinkType = "aws_s3"
	SinkTypeAzureLogsIngestion SinkType = "azure_logs_ingestion"
	SinkTypeAzureMonitorLogs   SinkType = "azure_monitor_logs"
	SinkTypeElasticsearch      SinkType = "elasticsearch"
	SinkTypeGcpStackdriverLogs SinkType = "gcp_stackdriver_logs"
	SinkTypeHttp               SinkType = "http"
	SinkTypeLoki               SinkType = "loki"
	SinkTypeKafka              SinkType = "kafka"
	SinkTypeOpenTelemetry      SinkType = "opentelemetry"
	SinkTypePrometheusExporter SinkType = "prometheus_exporter"
	SinkTypeSocket             SinkType = "socket"
	SinkTypeSplunkHecLogs      SinkType = "splunk_hec_logs"
)

type Sink interface {
	SinkType() SinkType
}
