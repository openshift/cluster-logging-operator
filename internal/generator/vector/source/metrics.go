package source

import "github.com/openshift/cluster-logging-operator/internal/generator/framework"

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

func (i InternalMetrics) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "internal_metrics"
{{end}}
`
}

func MetricsSources(id string) []framework.Element {
	return []framework.Element{
		InternalMetrics{
			ID:                id,
			ScrapeIntervalSec: 2,
		},
	}
}
