// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
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
		log.Error(err, "Unable to create collector config generator")
		os.Exit(1)
	}

	file := os.Args[1]
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error(err, "Error reading file", "file", file)
		os.Exit(1)
	}
	forwarder := &logging.ClusterLogForwarder{}
	err = yaml.Unmarshal(content, forwarder)
	if err != nil {
		log.Error(err, "Error Unmarshalling file ", "file", file)
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
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
