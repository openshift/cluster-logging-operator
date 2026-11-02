package types

type ExporterType string

const (
	ExporterTypeOtlpHttp ExporterType = "otlphttp"
)

// Exporter is an OpenTelemetry Collector exporter for sending signals out of the collector
type Exporter interface {
	ID() string
	ExporterType() ExporterType
}
