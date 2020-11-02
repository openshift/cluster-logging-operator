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

package fluentd

import (
	"fmt"
	"runtime"

	. "github.com/onsi/ginkgo"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[Collection] Provides only a fluentd daemonset", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		e2e        = helpers.NewE2ETestFramework()
		components []helpers.LogComponentType
	)

	Describe("when ClusterLogging is configured only with a collection spec", func() {

		BeforeEach(func() {
			if err := e2e.DeployLogGenerator(); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}

			components = []helpers.LogComponentType{helpers.ComponentTypeCollector}
			if err := e2e.SetupClusterLogging(components...); err != nil {
				Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
			}

		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluentd"})
		}, helpers.DefaultCleanUpTimeout)

		It("should default to a running collector", func() {
			for _, component := range components {
				if err := e2e.WaitFor(component); err != nil {
					Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
				}
			}
		})
	})
})
