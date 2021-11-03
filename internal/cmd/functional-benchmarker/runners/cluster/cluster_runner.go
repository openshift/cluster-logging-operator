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
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"path"
	"strings"
	"time"
)

const containerVolumeName = "container"

type ClusterRunner struct {
	framework *functional.FluentdFunctionalFramework
	config.Options
	loaders []loader
}

type loader struct {
	Name string
	File string
}

func New(options config.Options) *ClusterRunner {
	return &ClusterRunner{
		Options: options,
		loaders: []loader{},
	}
}

/* #nosec G306*/
func (r *ClusterRunner) DumpLoaderArtifacts() {
	maxDuration, _ := time.ParseDuration("5m")
	defaultRetryInterval, _ := time.ParseDuration("10s")
	for _, l := range r.loaders {
		var result string
		name := l.Name
		file := l.File
		err := wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
			result, err = r.framework.RunCommand(name, "cat", file)
			if result != "" && err == nil {
				return true, nil
			}
			log.V(4).Error(err, "Polling logs")
			return false, nil
		})
		if err == nil {
			file := path.Base(l.File)
			if err := ioutil.WriteFile(path.Join(r.ArtifactDir, file), []byte(result), 0655); err != nil {
				log.V(0).Error(err, "Unable to write l logs", "file", file)
			}
		}
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
	r.framework = functional.NewFluentdFunctionalFrameworkUsing(&testclient.Test, testclient.Close, r.Verbosity)
	r.framework.Conf = r.CollectorConfig

	functional.NewClusterLogForwarderBuilder(r.framework.Forwarder).
		FromInput(logging.InputNameApplication).
		ToFluentForwardOutput()
	err := r.framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
		func(b *runtime.PodBuilder) error {
			for i := 0; i < r.TotalLogStressors; i++ {
				name := fmt.Sprintf("loader-%d", i)
				file := fmt.Sprintf("%s/%s_%s_%s-12345.log", config.ContainerLogDir, b.Pod.Name, b.Pod.Namespace, name)
				args := []string{
					"generate",
					"--destination=file",
					"--output-format=crio",
					fmt.Sprintf("--source=%s", r.PayloadSource),
					fmt.Sprintf("--log-lines-rate=%d", r.LinesPerSecond),
					fmt.Sprintf("--file=%s", file),
				}
				if r.PayloadSource == "synthetic" {
					args = append(args, fmt.Sprintf("--synthetic-payload-size=%d", r.MsgSize))
				}

				b.AddContainer(name, config.LogStressorImage).
					WithCmdArgs(args).
					AddVolumeMount(containerVolumeName, "/var/log/containers", "", false).
					End()
				r.loaders = append(r.loaders, loader{Name: name, File: file})
			}
			b.AddEmptyDirVolume(containerVolumeName)
			b.GetContainer(constants.CollectorName).
				AddVolumeMount(containerVolumeName, "/var/log/containers", "", false).
				Update()
			return r.framework.AddBenchmarkForwardOutput(b, r.framework.Forwarder.Spec.Outputs[0])
		},
	})
	if err != nil {
		log.Error(err, "Error deploying test pod")
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
		log.V(3).Info("Sample Collector", "result", result)
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
