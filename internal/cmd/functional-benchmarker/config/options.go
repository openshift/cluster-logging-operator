package config

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test"
)

const (
	LogStressorImage = "quay.io/openshift-logging/cluster-logging-load-client:latest"
	ContainerLogDir  = "/var/log/pods"
)

type Options struct {
	Image               string
	TotalMessages       int
	MsgSize             int
	Verbosity           int
	BaseLine            bool
	DoCleanup           bool
	Sample              bool
	Platform            string
	Output              string
	TotalLogStressors   int
	LinesPerSecond      int
	ArtifactDir         string
	CollectorConfigPath string
	CollectorConfig     string
	ReadTimeout         string
	RunDuration         string
	SampleDuration      string
	PayloadSource       string
}

func InitOptions(logger logr.Logger) Options {
	options := Options{
		ReadTimeout: test.SuccessTimeout().String(),
	}
	fs := flag.NewFlagSet("functional-benchmarker", flag.ExitOnError)

	fs.StringVar(&options.Image, "image", "quay.io/openshift-logging/fluentd:1.14.5", "The Image to use to run the benchmark")
	//fs.IntVar(&options.TotalMessages, "tot-messages", 10000, "The number of messages to write per stressor")
	fs.IntVar(&options.MsgSize, "size", 1024, "The message size in bytes per stressor for 'synthetic' payload")
	fs.IntVar(&options.LinesPerSecond, "lines-per-sec", 1, "The log lines per second per stressor")
	fs.IntVar(&options.Verbosity, "verbosity", 0, "The output log level")
	fs.BoolVar(&options.DoCleanup, "do-cleanup", true, "set to false to preserve the namespace")
	fs.BoolVar(&options.BaseLine, "baseline", false, "run the test with a baseline config. This supercedes --collector-config")
	fs.BoolVar(&options.Sample, "sample", false, "set to true to dump a Sample message")
	//fs.StringVar(&options.Platform, "platform", "cluster", "The runtime environment: cluster, local. local requires podman")
	fs.StringVar(&options.PayloadSource, "payload-source", "synthetic", "The load message profile: synthetic,application,simple")

	fs.StringVar(&options.ReadTimeout, "read-timeout", test.SuccessTimeout().String(), "The read timeout duration to wait for logs")
	fs.StringVar(&options.RunDuration, "run-duration", "5m", "The duration of the test run")
	fs.StringVar(&options.SampleDuration, "sample-duration", "1s", "The frequency to sample cpu and memory")

	fs.IntVar(&options.TotalLogStressors, "tot-stressors", 1, "Total log stressors")
	fs.StringVar(&options.CollectorConfigPath, "collector-config", "", "The path to the collector config to use")
	fs.StringVar(&options.ArtifactDir, "artifact-dir", "", "The directory to write artifacts (default: Time.now())")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error parsing argument: %v", err)
		os.Exit(1)
	}

	logger.V(1).Info("Starting functional benchmarker", "args", options)

	if err := os.Setenv(constants.FluentdImageEnvVar, options.Image); err != nil {
		logger.Error(err, "Error setting fluent Image env var")
		os.Exit(1)
	}

	return options
}

func ReadConfig(configFile string, baseline bool, logger logr.Logger) string {
	if baseline {
		logger.V(0).Info("Using the baseline config")
		return FluentdBaselineConf
	}
	var reader func() ([]byte, error)
	switch configFile {
	case "-":
		logger.V(1).Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	case "":
		logger.V(1).Info("received empty configFile. Generating from CLF")
		return ""
	default:
		logger.V(1).Info("reading configfile", "filename", configFile)
		reader = func() ([]byte, error) { return ioutil.ReadFile(configFile) }
	}
	content, err := reader()
	if err != nil {
		logger.Error(err, "Error reading config")
		os.Exit(1)
	}
	return string(content)
}

func MustParseDuration(durationString, optionName string, logger logr.Logger) time.Duration {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		logger.Error(err, "Unable to parse duration", "option", optionName)
		os.Exit(1)
	}
	return duration
}
