package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
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

func (r *ClusterRunner) Config() string {
	return r.framework.Conf
}

func (r *ClusterRunner) Namespace() string {
	return r.framework.Namespace
}

func (r *ClusterRunner) Pod() string {
	return r.framework.Pod.Name
}

func (r *ClusterRunner) Deploy() {
	testclient := client.NewNamespaceClient()
	r.framework = functional.NewCollectorFunctionalFrameworkUsing(&testclient.Test, testclient.Close, r.Verbosity, logging.LogCollectionTypeFluentd)
	r.framework.Conf = r.CollectorConfig

	functional.NewClusterLogForwarderBuilder(r.framework.Forwarder).
		FromInput(logging.InputNameApplication).
		ToFluentForwardOutput()

	//modify config to only collect loader containers
	r.framework.VisitConfig = func(conf string) string {
		pattern := fmt.Sprintf("/var/log/pods/%s_*/loader-*/*.log", r.framework.Namespace)
		conf = strings.Replace(conf, "/var/log/pods/*/*/*.log", pattern, 1)
		conf = strings.Replace(conf, "/var/log/pods/**/*.log", pattern, 1)
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
					End()
			}
			b.AddHostPathVolume(containerVolumeName, constants.ContainerLogDir)
			b.AddHostPathVolume(PodLogsDirName, constants.PodLogDir)
			collectorBuilder := b.GetContainer(constants.CollectorName).
				AddVolumeMount(containerVolumeName, constants.ContainerLogDir, "", true).
				AddVolumeMount(PodLogsDirName, constants.PodLogDir, "", true).
				AddEnvVarFromFieldRef("NAMESPACE", "metadata.namespace").
				AddEnvVar("RUBY_GC_HEAP_OLDOBJECT_LIMIT_FACTOR", "0.9").
				WithPrivilege()
			if r.RequestCPU != "" {
				collectorBuilder.ResourceRequirements(corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU: resource.MustParse(r.RequestCPU),
					},
				})
			}
			collectorBuilder.Update()

			return r.framework.AddBenchmarkForwardOutput(b, r.framework.Forwarder.Spec.Outputs[0])
		},
	})
	if err != nil {
		log.Error(err, "Error deploying test pod", "pod", r.framework.Pod)
		os.Exit(1)
	}

}

func (r *ClusterRunner) ReadApplicationLogs() ([]string, error) {

	artifacts, err := ioutil.ReadDir(r.ArtifactDir)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, file := range artifacts {
		if strings.HasPrefix(file.Name(), "kubernetes.") {
			files = append(files, file.Name())
		}
	}
	logs := []string{}
	for _, file := range files {
		filePath := path.Join(r.ArtifactDir, file)
		result, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Error(err, "Trying to read application logs", "path", file)
		}
		appLogs := strings.Split(strings.TrimSpace(string(result)), "\n")
		log.V(4).Info("App logs", "file", file, "logs", appLogs)
		logs = append(logs, appLogs...)
	}
	log.V(3).Info("Returning all app logs", "logs", logs)
	return logs, nil
}
func (r *ClusterRunner) FetchApplicationLogs() error {
	out, err := oc.Exec().WithNamespace(r.framework.Namespace).Pod(r.framework.Name).Container(logging.OutputTypeFluentdForward).
		WithCmd("ls", "/tmp").Run()
	if err != nil {
		return err
	}
	files := []string{}
	for _, file := range strings.Split(out, "\n") {
		if strings.HasPrefix(file, "kubernetes.") && strings.HasSuffix(file, "log.log") {
			files = append(files, file)
		}
	}
	for _, file := range files {
		cmd := fmt.Sprintf("oc cp %s/%s:/tmp/%s %s/%s -c %s  --request-timeout=3m", r.framework.Namespace, r.framework.Name, file,
			r.ArtifactDir, file, strings.ToLower(logging.OutputTypeFluentdForward))
		log.V(2).Info("copy command", "cmd", cmd)
		out, err := oc.Literal().From(cmd).Run()
		if err != nil {
			log.V(2).Error(err, "Trying to retrieve application logs", "path", file, "out", out)
		}
	}
	return nil
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
