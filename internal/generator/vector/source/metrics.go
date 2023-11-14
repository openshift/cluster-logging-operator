package source

const (
	InternalMetricsSourceName = "internal_metrics"
)

type InternalMetrics struct {
	ID                string
	ScrapeIntervalSec int
}

func (InternalMetrics) Name() string {
	return "internalMetricsTemplate"
}

// #namespace = "collector"
// #scrape_interval_secs = {{.ScrapeIntervalSec}}
func (i InternalMetrics) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "internal_metrics"
{{end}}
`
}
