package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/go-logr/logr"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"

	"github.com/spf13/pflag"

	"github.com/ViaQ/logerr/v2/log"
)

// HACK - This command is for development use only
func main() {
	logLevel, present := os.LookupEnv("LOG_LEVEL")

	var logger logr.Logger
	if present {
		verbosity, err := strconv.Atoi(logLevel)
		if err != nil {
			log.NewLogger("cluster-logging-operator").Error(err, "LOG_LEVEL must be an integer")
			os.Exit(1)
		}
		logger = log.NewLogger("cluster-logging-operator", log.WithVerbosity(verbosity))

	} else {
		logger = log.NewLogger("cluster-logging-operator")
	}

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	includeDefaultLogStore := flag.Bool("include-default-store", true, "Include the default storage when generating the config")
	debugOutput := flag.Bool("debug-output", false, "Generate config normally, but replace output plugins with @stdout plugin, so that records can be printed in collector logs.")
	colltype := flag.String("collector", "fluentd", "collector type: fluentd or vector")
	help := flag.Bool("help", false, "This message")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if *help {
		pflag.Usage()
		os.Exit(1)
	}
	logCollectorType := logging.LogCollectionType(*colltype)
	if logCollectorType != logging.LogCollectionTypeFluentd && logCollectorType != logging.LogCollectionTypeVector {
		pflag.Usage()
		os.Exit(1)
	}
	logger.V(1).Info("Forwarder Generator Main", "Args", os.Args)

	var reader func() ([]byte, error)
	switch *yamlFile {
	case "-":
		logger.Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	case "":
		logger.Info("received empty yamlfile")
		reader = func() ([]byte, error) { return []byte{}, nil }
	default:
		logger.Info("reading log forwarder from yaml file", "filename", *yamlFile)
		reader = func() ([]byte, error) { return ioutil.ReadFile(*yamlFile) }
	}

	content, err := reader()
	if err != nil {
		logger.Error(err, "Error Unmarshalling file ", "file", yamlFile)
		os.Exit(1)
	}

	logger.Info("Finished reading yaml", "content", string(content))

	generatedConfig, err := forwarder.Generate(logCollectorType, string(content), *includeDefaultLogStore, *debugOutput, nil)
	if err != nil {
		logger.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
