package cluster

import (
	"fmt"
	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"os"
	"strings"
	"time"
)

const (
	containerVolumeName = "container"
	PodLogsDirName      = "varlogpods"
)

type ClusterRunner struct {
	framework *functional.CollectorFunctionalFramework
	config.Options
}

func New(options config.Options) *ClusterRunner {
	return &ClusterRunner{
		Options: options,
	}
}

func (r *ClusterRunner) Namespace() string {
	return r.framework.Namespace
}

func (r *ClusterRunner) Pod() string {
	return r.framework.Pod.Name
}

func (r *ClusterRunner) Deploy() {
	testclient := client.NewNamesapceClient()
	r.framework = functional.NewCollectorFunctionalFrameworkUsing(&testclient.Test, testclient.Close, r.Verbosity)
	r.framework.Conf = r.CollectorConfig

	functional.NewClusterLogForwarderBuilder(r.framework.Forwarder).
		FromInput(logging.InputNameApplication).
		ToFluentForwardOutput()

	//modify config to only collect loader containers
	r.framework.VisitConfig = func(conf string) string {
		conf = strings.Replace(conf, "/var/log/containers/*.log", "/var/log/containers/*_*_loader-*.log", 1)
		return conf
	}

	err := r.framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
		func(b *runtime.PodBuilder) error {
			for i := 0; i < r.TotalLogStressors; i++ {
				name := fmt.Sprintf("loader-%d", i)
				args := []string{
					"generate",
					fmt.Sprintf("--source=%s", r.PayloadSource),
					fmt.Sprintf("--log-lines-rate=%d", r.LinesPerSecond),
				}
				if r.PayloadSource == "synthetic" {
					args = append(args, fmt.Sprintf("--synthetic-payload-size=%d", r.MsgSize))
				}

				b.AddContainer(name, config.LogStressorImage).
					WithCmdArgs(args).
					AddVolumeMount(containerVolumeName, constants.ContainerLogDir, "", false).
					AddVolumeMount(PodLogsDirName, constants.PodLogDir, "", false).
					End()
			}
			b.AddHostPathVolume(containerVolumeName, constants.ContainerLogDir)
			b.AddHostPathVolume(PodLogsDirName, constants.PodLogDir)
			collectorBuilder := b.GetContainer(constants.CollectorName).
				AddVolumeMount(containerVolumeName, constants.ContainerLogDir, "", true).
				AddVolumeMount(PodLogsDirName, constants.PodLogDir, "", true).
				WithPrivilege()
			collectorBuilder.Update()
			return r.framework.AddBenchmarkForwardOutput(b, r.framework.Forwarder.Spec.Outputs[0])
		},
	})
	if err != nil {
		log.Error(err, "Error deploying test pod", "pod", r.framework.Pod)
		os.Exit(1)
	}

}
func (r *ClusterRunner) WritesApplicationLogsOfSize(msgSize int) error {
	return r.framework.WritesNApplicationLogsOfSize(r.TotalMessages, msgSize)
}

func (r *ClusterRunner) ReadApplicationLogs() ([]string, error) {
	return r.framework.ReadRawApplicationLogsFrom(logging.OutputTypeFluentdForward)
}
func (r *ClusterRunner) Cleanup() {
	r.framework.Cleanup()
}

func (r *ClusterRunner) SampleCollector() *stats.Sample {
	if result, err := oc.AdmTop(r.framework.Namespace, r.framework.Name).NoHeaders().ForContainers().Run(); err == nil {
		log.V(3).Info("Sample collector", "result", result)
		if !strings.Contains(result, "Error from server") {
			for _, line := range strings.Split(result, "\n") {
				fields := strings.Fields(line)
				log.V(3).Info("Container metric", "fields", fields)
				if len(fields) == 4 && fields[1] == constants.CollectorName {
					return &stats.Sample{
						Time:        time.Now().Unix(),
						CPUCores:    fields[2],
						MemoryBytes: fields[3],
					}
				}
			}
		}

	} else {
		log.V(3).Error(err, "Unable to sample collector metrics", "result", result)
	}
	return nil
}
