package main

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/reports"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/runners/cluster"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ViaQ/logerr/log"
)

// HACK - This command is for development use only
func main() {
	options := config.InitOptions()

	options.CollectorConfig = config.ReadConfig(options.CollectorConfigPath)
	log.V(1).Info(options.CollectorConfig)

	artifactDir := createArtifactDir(options.ArtifactDir)

	metrics, statistics, err := RunBenchmark(artifactDir, options)
	if err != nil {
		log.Error(err, "Error in run")
		os.Exit(1)
	}

	reporter := reports.NewReporter(options.Output, artifactDir, metrics, statistics)
	reporter.Generate()
}

func RunBenchmark(artifactDir string, options config.Options) (*stats.ResourceMetrics, *stats.Statistics, error) {
	runDuration := config.MustParseDuration(options.RunDuration, "run-duration")
	sampleDuration := config.MustParseDuration(options.SampleDuration, "resource-sample-duration")
	runner := NewRunner(options)
	runner.Deploy()
	if options.DoCleanup {
		log.V(2).Info("Deferring cleanup", "DoCleanup", options.DoCleanup)
		defer runner.Cleanup()
	}
	//delay waiting to spin up
	log.V(1).Info("delaying awaiting pod start")
	time.Sleep(1 * time.Minute)
	done := make(chan bool)
	startTime := time.Now()
	sampler := time.NewTicker(sampleDuration)
	metrics := stats.NewResourceMetrics()
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
	statistics := gatherStatistics(runner, options.Sample, options.MsgSize, startTime, endTime)
	done <- true
	sampler.Stop()

	return metrics, statistics, nil
}

func gatherStatistics(runner Runner, sample bool, msgSize int, startTime, endTime time.Time) *stats.Statistics {
	raw, err := runner.ReadApplicationLogs()
	if err != nil {
		log.Error(err, "Error reading logs")
		return nil
	}
	log.V(4).Info("Read logs", "raw", raw)
	logs := types.PerfLogs{}
	err = json.Unmarshal([]byte(utils.ToJsonLogs(raw)), &logs)
	if err != nil {
		log.Error(err, "Error parsing logs")
		return nil
	}
	log.V(4).Info("Read logs", "parsed", logs)
	if sample {
		fmt.Printf("Sample:\n%s\n", test.JSONString(logs[0]))
	}
	log.V(4).Info("Removing logs outside the duration")
	filtered := types.PerfLogs{}
	for _, entry := range logs {
		if entry.Timestamp.After(startTime) && entry.Timestamp.Before(endTime) {
			filtered = append(filtered, entry)
		}
	}
	log.V(4).Info("filtered", "logs", filtered)
	return stats.NewStatisics(filtered, msgSize, endTime.Sub(startTime))
}

type Runner interface {
	Deploy()
	WritesApplicationLogsOfSize(msgSize int) error
	ReadApplicationLogs() ([]string, error)
	SampleCollector() *stats.Sample
	Cleanup()
}

func NewRunner(options config.Options) Runner {
	return &cluster.ClusterRunner{
		Options: options,
	}
}

func createArtifactDir(rootDir string) string {
	var err error
	artifactDir := path.Join(rootDir, fmt.Sprintf("benchmark-%s", time.Now().Format(time.RFC3339)))
	if err = os.Mkdir(artifactDir, 0755); err != nil {
		log.Error(err, "Error creating temp director")
		os.Exit(1)
	}
	if err := os.Chmod(artifactDir, 0755); err != nil {
		log.Error(err, "Error modifying temp director permissions")
		os.Exit(1)
	}
	if artifactDir, err = filepath.Abs(artifactDir); err != nil {
		log.Error(err, "Unable to determine the absolute file path of the artifactDir")
		os.Exit(1)
	}
	return artifactDir
}
