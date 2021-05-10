package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ViaQ/logerr/log"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

type Options struct {
	image               string
	totalMessages       int
	msgSize             int
	verbosity           int
	doCleanup           bool
	sample              bool
	platform            string
	output              string
	totalLogStressors   int
	collectorConfigPath string
	readTimeout         string
}

// HACK - This command is for development use only
func main() {
	options := Options{
		readTimeout: test.SuccessTimeout().String(),
	}
	fs := flag.NewFlagSet("functional-benchmarker", flag.ExitOnError)

	fs.StringVar(&options.image, "image", "quay.io/openshift/origin-logging-fluentd:latest", "The image to use to run the benchmark")
	fs.IntVar(&options.totalMessages, "totMessages", 10000, "The number of messages to write per stressor")
	fs.IntVar(&options.msgSize, "size", 1024, "The message size in bytes")
	fs.IntVar(&options.verbosity, "verbosity", 0, "")
	fs.BoolVar(&options.doCleanup, "docleanup", true, "set to false to preserve the namespace")
	fs.BoolVar(&options.sample, "sample", false, "set to true to dump a sample message")
	fs.StringVar(&options.platform, "platform", "cluster", "The runtime environment: cluster (default), local (experimental). local requires podman")
	fs.StringVar(&options.output, "output", "default", "The output format: default, csv")
	fs.StringVar(&options.readTimeout, "readTimeout", test.SuccessTimeout().String(), "The read timeout duration to wait for logs")

	fs.IntVar(&options.totalLogStressors, "exp-tot-stressors", 1, "Experimental. Total log stressors (platform=local)")
	fs.StringVar(&options.collectorConfigPath, "exp-collector-config", "", "Experimental. The collector config to use (platform=local")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error parsing argument: %v", err)
		os.Exit(1)
	}

	log.MustInit("functional-benchmark")
	log.SetLogLevel(options.verbosity)
	log.V(1).Info("Starting functional benchmarker", "args", options)

	if err := os.Setenv(constants.FluentdImageEnvVar, options.image); err != nil {
		log.Error(err, "Error setting fluent image env var")
		os.Exit(1)
	}

	collectorConfig := ReadConfig(options.collectorConfigPath)
	log.V(1).Info(collectorConfig)

	runs := map[string]string{
		"baseline": fluentdBaselineConf,
		"config":   collectorConfig,
	}

	reporter := NewReporter(options.output)
	for name, config := range runs {
		log.V(1).Info("Executing", "run", name)
		stats, metrics, err := NewRun(config, options)()
		if err != nil {
			log.Error(err, "Error in run")
			os.Exit(1)
		}
		reporter.Add(name, *stats, *metrics)
	}
	reporter.Print()
}
func NewRun(collectorConfig string, options Options) func() (*Statistics, *Metrics, error) {
	return func() (*Statistics, *Metrics, error) {
		runner := NewBencharker(collectorConfig, options)
		runner.Deploy()
		if options.doCleanup {
			log.V(2).Info("Deferring cleanup", "doCleanup", options.doCleanup)
			defer runner.Cleanup()
		}

		startTime := time.Now()
		var (
			logs    []string
			readErr error
			metrics Metrics
		)
		done := make(chan bool)
		go func() {
			logs, readErr = runner.ReadApplicationLogs()
			metrics = runner.Metrics()
			done <- true
		}()
		//defer reader to get logs
		if err := runner.WritesApplicationLogsOfSize(options.msgSize); err != nil {
			return nil, nil, fmt.Errorf("Error writing application logs: %v", err)
		}
		<-done
		endTime := time.Now()
		if readErr != nil {
			return nil, nil, fmt.Errorf("Error reading application logs: %v", readErr)
		}
		log.V(4).Info("Read logs", "raw", logs)
		perflogs := types.PerfLogs{}
		err := json.Unmarshal([]byte(utils.ToJsonLogs(logs)), &perflogs)
		if err != nil {
			return nil, nil, fmt.Errorf("Error parsing logs: %v", err)
		}
		log.V(4).Info("Read logs", "parsed", perflogs)
		log.V(4).Info("Read logs", "parsed", perflogs)
		if options.sample {
			fmt.Printf("Sample:\n%s\n", test.JSONString(perflogs[0]))
		}
		return NewStatisics(perflogs, options.msgSize, endTime.Sub(startTime)), &metrics, nil
	}
}

type Runner interface {
	Deploy()
	WritesApplicationLogsOfSize(msgSize int) error
	ReadApplicationLogs() ([]string, error)
	Metrics() Metrics
	Cleanup()
}

func NewBencharker(collectorConfig string, options Options) Runner {
	if options.platform == "local" {
		return NewLocalRunner(collectorConfig, options)
	}
	return &ClusterRunner{verbosity: options.verbosity, totalMessages: options.totalMessages}
}

func ReadConfig(configFile string) string {
	var reader func() ([]byte, error)
	switch configFile {
	case "-":
		log.V(1).Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	case "":
		log.V(1).Info("received empty configFile. Generating from CLF")
		return ""
	default:
		log.V(1).Info("reading configfile", "filename", configFile)
		reader = func() ([]byte, error) { return ioutil.ReadFile(configFile) }
	}
	content, err := reader()
	if err != nil {
		log.Error(err, "Error reading config")
		os.Exit(1)
	}
	return string(content)
}

type Metrics struct {
	cpuUserTicks     string
	cpuKernelTicks   string
	memVirtualPeakKB string
}
