package runners

import (
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/runners/cluster"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
)

type Runner interface {
	Deploy()
	WritesApplicationLogsOfSize(msgSize int) error
	ReadApplicationLogs() ([]string, error)
	SampleCollector() *stats.Sample
	Cleanup()
	Namespace() string
	Pod() string
	DumpLoaderArtifacts()
}

func NewRunner(options config.Options) Runner {
	return cluster.New(options)
}
