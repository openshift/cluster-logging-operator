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

package forwarding

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentd"
)

//NewConfigGenerator create a config generator for a given collector type
func NewConfigGenerator(collector logging.LogCollectionType, includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin bool) (ConfigGenerator, error) {
	switch collector {
	case logging.LogCollectionTypeFluentd:
		return fluentd.NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin)
	default:
		return nil, fmt.Errorf("Config generation not supported for collectors of type %s", collector)
	}
}

//ConfigGenerator is a config generator for a given ClusterLogForwarderSpec
type ConfigGenerator interface {

	//Generate the config
	Generate(clfSpec *logging.ClusterLogForwarderSpec, fwSpec *logging.ForwarderSpec) (string, error)
}
