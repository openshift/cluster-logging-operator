package vector

const (
	InternalMetricsSourceName = "internal_metrics"
	PrometheusOutputSinkName  = "prometheus_output"
	PrometheusExporterAddress = "[::]:24231"

	AddNodenameToMetricTransformName = "add_nodename_to_metric"
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

[sinks.{{.ID}}.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
{{end}}`
}

type AddNodenameToMetric struct {
	ID     string
	Inputs string
}

func (a AddNodenameToMetric) Name() string {
	return AddNodenameToMetricTransformName
}

func (a AddNodenameToMetric) Template() string {
	return `{{define "` + a.Name() + `" -}}
[transforms.{{.ID}}]
type = "remap"
inputs = {{.Inputs}}
source = '''
.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''
{{end}}`
}
