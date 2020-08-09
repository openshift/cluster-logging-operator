package main

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/sirupsen/logrus"
)

// HACK - This command is for development use only
func main() {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Errorf("Unable to evaluate the LOG_LEVEL: %s %v", logLevel, err)
			os.Exit(1)
		}
		logrus.SetLevel(level)
	}

	if len(os.Args) < 2 {
		logrus.Error("Need to pass the logging forwarding yaml as an arg")
		os.Exit(1)
	}

	logCollectorType := logging.LogCollectionTypeFluentd
	includeLegacyForward := false
	includeLegacySyslog := false
	useOldRemoteSyslogPlugin := false
	includeDefaultLogStore := true

	generator, err := forwarding.NewConfigGenerator(
		logCollectorType,
		includeLegacyForward,
		includeLegacySyslog,
		useOldRemoteSyslogPlugin)
	if err != nil {
		logger.Errorf("Unable to create collector config generator: %v", err)
		os.Exit(1)
	}

	yamlFile := os.Args[1]
	content, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		logger.Errorf("Error reading file %s: %v", yamlFile, err)
		os.Exit(1)
	}
	forwarder := &logging.ClusterLogForwarder{}
	err = yaml.Unmarshal(content, forwarder)
	if err != nil {
		logger.Errorf("Error Unmarshalling file %s: %v", yamlFile, err)
		os.Exit(1)
	}

	clRequest := &k8shandler.ClusterLoggingRequest{
		ForwarderRequest: forwarder,
	}
	if includeDefaultLogStore {
		clRequest.Cluster = &logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				LogStore: &logging.LogStoreSpec{
					Type: logging.LogStoreTypeElasticsearch,
				},
			},
		}
	}
	spec, _ := clRequest.NormalizeForwarder()
	tunings := &logging.ForwarderSpec{}

	generatedConfig, err := generator.Generate(spec, tunings)
	if err != nil {
		logger.Warnf("Unable to generate log configuration: %v", err)
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
