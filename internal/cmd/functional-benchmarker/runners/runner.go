package runners

import (
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/runners/cluster"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
)

type Runner interface {
	Deploy()
	ReadApplicationLogs() ([]string, error)
	FetchApplicationLogs() error
	SampleCollector() *stats.Sample
	Cleanup()
	Namespace() string
	Pod() string
	Config() string
}

func NewRunner(options config.Options) Runner {
	return cluster.New(options)
}
