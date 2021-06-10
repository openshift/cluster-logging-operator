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
		log.MustInit("cluster-logging-operator")
		log.SetLogLevel(verbosity)
	} else {
		log.MustInit("cluster-logging-operator")
	}

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	includeDefaultLogStore := flag.Bool("include-default-store", true, "Include the default storage when generating the config")
	includeLegacyForward := flag.Bool("include-legacy-forward", false, "Include the legacy forward when generating the config")
	help := flag.Bool("help", false, "This message")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if *help {
		pflag.Usage()
		os.Exit(1)
	}
	log.V(1).Info("Forwarder Generator Main", "Args", os.Args)

	var reader func() ([]byte, error)
	switch *yamlFile {
	case "-":
		log.Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	case "":
		log.Info("received empty yamlfile")
		reader = func() ([]byte, error) { return []byte{}, nil }
	default:
		log.Info("reading log forwarder from yaml file", "filename", *yamlFile)
		reader = func() ([]byte, error) { return ioutil.ReadFile(*yamlFile) }
	}

	content, err := reader()
	if err != nil {
		log.Error(err, "Error Unmarshalling file ", "file", yamlFile)
		os.Exit(1)
	}

	log.Info("Finished reading yaml", "content", string(content))

	generatedConfig, err := forwarder.Generate(string(content), *includeDefaultLogStore, *includeLegacyForward)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
