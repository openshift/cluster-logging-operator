package cluster

import (
	"bufio"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	obsruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
	testClient := client.NewNamespaceClient()
	cleanup := func() {
		if r.DoCleanup {
			testClient.Close()
		}
	}
	r.framework = functional.NewCollectorFunctionalFrameworkUsing(&testClient.Test, cleanup, r.Verbosity)
	r.framework.Conf = r.CollectorConfig

	obsruntime.NewClusterLogForwarderBuilder(r.framework.Forwarder).
		FromInputName("benchmark", func(spec *obs.InputSpec) {
			spec.Type = obs.InputTypeApplication
			spec.Application = &obs.Application{
				Includes: []obs.NamespaceContainerSpec{
					{Namespace: r.Namespace()},
				},
			}
		}).
		ToHttpOutput()

	err := r.framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
		func(b *runtime.PodBuilder) error {
			for i := 0; i < r.TotalLogStressors; i++ {
				name := fmt.Sprintf("loader-%d", i)
				args := []string{
					"--command=generate",
					"--use-random-hostname",
					fmt.Sprintf("--log-type=%s", r.PayloadSource),
					fmt.Sprintf("--logs-per-second=%d", r.LinesPerSecond),
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
				WithPrivilege()
			if r.RequestCPU != "" {
				collectorBuilder.ResourceRequirements(corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU: resource.MustParse(r.RequestCPU),
					},
				})
			}
			collectorBuilder.Update()
			return r.framework.AddBenchmarkForwardOutput(b, r.framework.Forwarder.Spec.Outputs[0], utils.GetComponentImage(constants.VectorName))
		},
	})
	if err != nil {
		log.Error(err, "Error deploying test pod", "pod", r.framework.Pod)
		os.Exit(1)
	}

}

func (r *ClusterRunner) ReadApplicationLogs() (stats.PerfLogs, error) {

	artifacts, err := os.ReadDir(r.ArtifactDir)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, file := range artifacts {
		if strings.HasPrefix(file.Name(), "loader-") {
			files = append(files, file.Name())
		}
	}
	mt := sync.Mutex{}
	logs := stats.PerfLogs{}
	wg := sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		filePath := path.Join(r.ArtifactDir, file)
		go func() {
			defer wg.Done()
			entries, err := ReadAndParseFile(filePath)
			if err != nil {
				log.Error(err, "Trying to read application logs", "path", filePath)
			}
			log.V(4).Info("App logs", "file", filePath, "logs", entries)
			defer mt.Unlock()
			mt.Lock()
			logs = append(logs, entries...)
		}()
	}
	wg.Wait()
	log.V(3).Info("Returning all app logs", "logs", logs)
	return logs, nil
}

func ReadAndParseFile(filePath string) (stats.PerfLogs, error) {
	log.V(4).Info("Reading and parsing file", "file", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err, "Unable to open file for analysis", "file", filePath)
	}
	defer file.Close()

	entries := stats.PerfLogs{}
	scanner := bufio.NewScanner(file)
	purged := 0
	for scanner.Scan() {
		if entry := stats.NewPerfLog(scanner.Text()); entry != nil {
			entries = append(entries, *entry)
		} else {
			purged += 1
		}
	}

	if purged > 0 {
		log.V(0).Info("Purged entries while parsing results", "purged", purged, "file", filePath)
	}

	if err := scanner.Err(); err != nil {
		log.Error(err, "Failed scanning file for analysis", "file", filePath)
		return nil, err
	}
	return entries, nil
}

func (r *ClusterRunner) FetchApplicationLogs() error {
	log.V(3).Info("Fetching Application Logs ...")
	out, err := oc.Exec().WithNamespace(r.framework.Namespace).Pod(r.framework.Name).Container(string(obs.OutputTypeHTTP)).
		WithCmd("ls", "/tmp").Run()
	if err != nil {
		return err
	}
	log.V(3).Info("Received Application logs", "files", out)
	files := []string{}
	for _, file := range strings.Split(out, "\n") {
		if strings.HasPrefix(file, "loader-") {
			files = append(files, file)
		}
	}
	for _, file := range files {
		cmd := fmt.Sprintf("oc cp %s/%s:/tmp/%s %s/%s -c %s  --request-timeout=3m", r.framework.Namespace, r.framework.Name, file,
			r.ArtifactDir, file, strings.ToLower(string(obs.OutputTypeHTTP)))
		log.V(2).Info("copy command", "cmd", cmd)
		out, err := oc.Literal().From(cmd).Run()
		if err != nil {
			log.V(2).Error(err, "Trying to retrieve application logs", "path", file, "out", out)
		}
	}
	return nil
}

func (r *ClusterRunner) Cleanup() {
	if r.Options.DoCleanup {
		r.framework.Cleanup()
	}
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
