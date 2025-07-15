package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"

	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/reports"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/runners"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

var (
	log = utils.InitStaticLogger("functional-benchmarker")
)

// HACK - This command is for development use only
func main() {
	defer func() {
		log.Info("Test complete")
	}()
	options := config.InitOptions()

	options.CollectorConfig = config.ReadConfig(options.CollectorConfigPath, options.BaseLine)
	log.V(1).Info(options.CollectorConfig)

	artifactDir := createArtifactDir(options.ArtifactDir)
	if options.ArtifactDir != artifactDir {
		options.ArtifactDir = artifactDir
	}

	metrics, statistics, config, err := RunBenchmark(artifactDir, options)
	if err != nil {
		log.Error(err, "Error in run")
		os.Exit(1)
	}

	options.CollectorConfig = config
	reporter := reports.NewReporter(options, artifactDir, metrics, statistics)
	log.Info("Generating reports...")
	reporter.Generate()
}

func RunBenchmark(artifactDir string, options config.Options) (*stats.ResourceMetrics, *stats.Statistics, string, error) {
	runDuration := config.MustParseDuration(options.RunDuration, "run-duration")
	sampleDuration := config.MustParseDuration(options.SampleDuration, "resource-sample-duration")
	runner := runners.NewRunner(options)
	log.Info("Deploying collector, loaders, and receivers")
	runner.Deploy()
	if options.DoCleanup {
		log.V(2).Info("Deferring cleanup", "DoCleanup", options.DoCleanup)
		defer runner.Cleanup()
	}
	done := make(chan bool)
	startTime := time.Now()
	sampler := time.NewTicker(sampleDuration)
	metrics := stats.NewResourceMetrics()
	log.V(0).Info("Starting to sample metrics for the collector...")
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				{
					log.V(3).Info("Collecting Sample")
					metrics.AddSample(runner.SampleCollector())
				}

			}
		}
	}()
	time.Sleep(runDuration)
	endTime := time.Now()
	done <- true
	sampler.Stop()
	log.Info("Stopped sampling metrics")
	log.Info("Fetching log data from receiver...")
	if err := runner.FetchApplicationLogs(); err != nil {
		return nil, nil, "", err
	}
	statistics := gatherStatistics(runner, options.Sample, options.MsgSize, startTime, endTime)

	fetchContainerLogs(runner, artifactDir)

	return metrics, statistics, runner.Config(), nil
}

func fetchContainerLogs(runner runners.Runner, artifactDir string) {
	var err error
	var out string
	for _, container := range []string{constants.CollectorName, string(obs.OutputTypeHTTP)} {

		if out, err = oc.Logs().WithNamespace(runner.Namespace()).WithPod(runner.Pod()).WithContainer(strings.ToLower(container)).Run(); err == nil {

			/* #nosec G306*/
			err = os.WriteFile(path.Join(artifactDir, container+".logs"), []byte(out), 0755)
			if err != nil {
				log.Error(err, "Error writing collector logs to artifactDir")
			}
		} else {
			log.Error(err, "Error retrieving collector logs from container", "container")
		}
	}
}

func gatherStatistics(runner runners.Runner, sample bool, msgSize int, startTime, endTime time.Time) *stats.Statistics {
	log.Info("Evaluating log data to calculate statistics")
	logs, err := runner.ReadApplicationLogs()
	if err != nil {
		log.Error(err, "Error reading logs")
		return &stats.Statistics{}
	}
	log.V(4).Info("Parsed logs", "parsed", logs)
	return stats.NewStatisics(logs, msgSize, endTime.Sub(startTime))
}

func createArtifactDir(artifactDir string) string {
	if strings.TrimSpace(artifactDir) == "" {
		artifactDir = fmt.Sprintf("./benchmark-%s", time.Now().Format(time.RFC3339Nano))
		artifactDir = strings.ReplaceAll(artifactDir, ":", "_")
	}
	var err error

	if err = os.Mkdir(artifactDir, 0755); err != nil {
		log.Error(err, "Error creating artifact directory")
		os.Exit(1)
	}
	if err := os.Chmod(artifactDir, 0755); err != nil {
		log.Error(err, "Error modifying artifact directory permissions")
		os.Exit(1)
	}
	if artifactDir, err = filepath.Abs(artifactDir); err != nil {
		log.Error(err, "Unable to determine the absolute file path of the artifact directory")
		os.Exit(1)
	}
	return artifactDir
}
