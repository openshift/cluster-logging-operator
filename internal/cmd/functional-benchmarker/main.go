package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/config"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/reports"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/runners"
	"github.com/openshift/cluster-logging-operator/internal/cmd/functional-benchmarker/stats"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
)

// HACK - This command is for development use only
func main() {
	logger := utils.InitLogger("functional-benchmarker")

	options := config.InitOptions(logger)

	options.CollectorConfig = config.ReadConfig(options.CollectorConfigPath, options.BaseLine, logger)
	logger.V(1).Info(options.CollectorConfig)

	artifactDir := createArtifactDir(options.ArtifactDir, logger)

	metrics, statistics, err := RunBenchmark(artifactDir, options, logger)
	if err != nil {
		logger.Error(err, "Error in run")
		os.Exit(1)
	}

	reporter := reports.NewReporter(options, artifactDir, metrics, statistics)
	reporter.Generate()
}

func RunBenchmark(artifactDir string, options config.Options, logger logr.Logger) (*stats.ResourceMetrics, *stats.Statistics, error) {
	runDuration := config.MustParseDuration(options.RunDuration, "run-duration", logger)
	sampleDuration := config.MustParseDuration(options.SampleDuration, "resource-sample-duration", logger)
	runner := runners.NewRunner(options)
	runner.Deploy()
	if options.DoCleanup {
		logger.V(2).Info("Deferring cleanup", "DoCleanup", options.DoCleanup)
		defer runner.Cleanup()
	}
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
					logger.V(3).Info("Collecting Sample")
					metrics.AddSample(runner.SampleCollector())
				}

			}
		}
	}()
	time.Sleep(runDuration)
	endTime := time.Now()
	done <- true
	statistics, _ := gatherStatistics(runner, options.Sample, options.MsgSize, startTime, endTime, logger)
	sampler.Stop()

	return metrics, statistics, nil
}

func gatherStatistics(runner runners.Runner, sample bool, msgSize int, startTime, endTime time.Time, logger logr.Logger) (*stats.Statistics, []string) {
	raw, err := runner.ReadApplicationLogs()
	if err != nil {
		logger.Error(err, "Error reading logs")
		return &stats.Statistics{}, []string{}
	}
	logger.V(4).Info("Read logs", "raw", raw)
	logs := stats.PerfLogs{}
	err = json.Unmarshal([]byte(utils.ToJsonLogs(raw)), &logs)
	if err != nil {
		logger.Error(err, "Error parsing logs")
		return &stats.Statistics{}, []string{}
	}
	logger.V(4).Info("Read logs", "parsed", logs)
	if sample {
		fmt.Printf("Sample:\n%s\n", test.JSONString(logs[0]))
	}
	return stats.NewStatisics(logs, msgSize, endTime.Sub(startTime)), raw
}

func createArtifactDir(artifactDir string, logger logr.Logger) string {
	if strings.TrimSpace(artifactDir) == "" {
		artifactDir = fmt.Sprintf("./benchmark-%s", time.Now().Format(time.RFC3339Nano))
	}
	var err error

	if err = os.Mkdir(artifactDir, 0755); err != nil {
		logger.Error(err, "Error creating artifact directory")
		os.Exit(1)
	}
	if err := os.Chmod(artifactDir, 0755); err != nil {
		logger.Error(err, "Error modifying artifact directory permissions")
		os.Exit(1)
	}
	if artifactDir, err = filepath.Abs(artifactDir); err != nil {
		logger.Error(err, "Unable to determine the absolute file path of the artifact directory")
		os.Exit(1)
	}
	return artifactDir
}
