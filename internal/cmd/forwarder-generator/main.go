package main

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/ViaQ/logerr/log"
)

// HACK - This command is for development use only
func main() {
	if len(os.Args) < 2 {
		log.Info("Need to pass the logging forwarding yaml as an arg")
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
		log.Error(err,"Unable to create collector config generator")
		os.Exit(1)
	}

	file := os.Args[1]
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error(err,"Error reading file", "file", file)
		os.Exit(1)
	}
	forwarder := &logging.ClusterLogForwarder{}
	err = yaml.Unmarshal(content, forwarder)
	if err != nil {
		log.Error(err,"Error Unmarshalling file ", "file", file)
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

	generatedConfig, err := generator.Generate(spec, nil, tunings)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
