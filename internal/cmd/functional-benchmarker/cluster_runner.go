package main

import (
	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"os"
)

type ClusterRunner struct {
	verbosity     int
	totalMessages int
	framework     *functional.FluentdFunctionalFramework
}

func (r *ClusterRunner) Deploy() {
	testclient := client.NewNamesapceClient()
	r.framework = functional.NewFluentdFunctionalFrameworkUsing(&testclient.Test, testclient.Close, r.verbosity)

	functional.NewClusterLogForwarderBuilder(r.framework.Forwarder).
		FromInput(logging.InputNameApplication).
		ToFluentForwardOutput()
	err := r.framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
		func(b *runtime.PodBuilder) error {
			return r.framework.AddBenchmarkForwardOutput(b, r.framework.Forwarder.Spec.Outputs[0])
		},
	})
	if err != nil {
		log.Error(err, "Error deploying test pod")
		os.Exit(1)
	}

}
func (r *ClusterRunner) WritesApplicationLogsOfSize(msgSize int) error {
	return r.framework.WritesNApplicationLogsOfSize(r.totalMessages, msgSize)
}

func (r *ClusterRunner) ReadApplicationLogs() ([]string, error) {
	return r.framework.ReadNApplicationLogsFrom(uint64(r.totalMessages), logging.OutputTypeFluentdForward)
}
func (r *ClusterRunner) Cleanup() {
	r.framework.Cleanup()
}

func (r *ClusterRunner) Metrics() Metrics {
	return Metrics{}
}
