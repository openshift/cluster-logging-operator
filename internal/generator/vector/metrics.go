package vector

const (
	InternalMetricsSourceName = "internal_metrics"
	PrometheusOutputSinkName  = "prometheus_output"
	PrometheusExporterAddress = "0.0.0.0:24231"
)

type InternalMetrics struct {
	ID                string
	ScrapeIntervalSec int
}

func (InternalMetrics) Name() string {
	return "internalMetricsTemplate"
}

//#namespace = "collector"
//#scrape_interval_secs = {{.ScrapeIntervalSec}}
func (i InternalMetrics) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "internal_metrics"
{{end}}
`
}

type PrometheusExporter struct {
	ID      string
	Inputs  string
	Address string
}

func (p PrometheusExporter) Name() string {
	return "PrometheusExporterTemplate"
}

func (p PrometheusExporter) Template() string {
	return `{{define "` + p.Name() + `" -}}
[sinks.{{.ID}}]
type = "prometheus_exporter"
inputs = {{.Inputs}}
address = "{{.Address}}"
default_namespace = "collector"
{{end}}`
}
