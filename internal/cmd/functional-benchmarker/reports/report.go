package reports

import (
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/reports/gnuplot"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
)

type Report interface {
	Generate()
}

func NewReporter(options config.Options, artifactDir string, metrics *stats.ResourceMetrics, statistics *stats.Statistics) Report {
	return &gnuplot.GNUPlotReporter{
		Options:     options,
		Metrics:     *metrics,
		Stats:       *statistics,
		ArtifactDir: artifactDir,
	}
}
