package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/spf13/pflag"

	"github.com/ViaQ/logerr/log"
)

// HACK - This command is for development use only
func main() {
	logLevel, present := os.LookupEnv("LOG_LEVEL")
	if present {
		verbosity, err := strconv.Atoi(logLevel)
		if err != nil {
			log.Error(err, "LOG_LEVEL must be an integer", "value", logLevel)
			os.Exit(1)
		}
		opt := log.WithVerbosity(uint8(verbosity))
		log.MustInitWithOptions("cluster-logging-operator", []log.Option{opt})
	} else {
		log.MustInit("cluster-logging-operator")
	}

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	includeDefaultLogStore := flag.Bool("include-default-store", true, "Include the default storage when generating the config")
	help := flag.Bool("help", false, "This message")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	log.V(1).Info("Args: %v", os.Args)

	if *help || len(os.Args) == 0 {
		pflag.Usage()
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		log.Info("Need to pass the logging forwarding yaml as an arg")
		os.Exit(1)
	}

	var reader func() ([]byte, error)
	if *yamlFile != "-" {
		reader = func() ([]byte, error) { return ioutil.ReadFile(*yamlFile) }
	} else {
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	}

	content, err := reader()
	if err != nil {
		log.Error(err, "Error Unmarshalling file ", "file", yamlFile)
		os.Exit(1)
	}

	generatedConfig, err := forwarder.Generate(string(content), *includeDefaultLogStore)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
