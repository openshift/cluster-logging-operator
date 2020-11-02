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

package sysloglegacy

import (
	"fmt"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[ClusterLogForwarder] Forwards logs", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		err              error
		syslogDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		testDir          string
	)
	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			log.Error(err, "unable to deploy log generator")
		}
		testDir = filepath.Dir(filename)
	})
	Describe("when the output in `syslog.conf` configmap is a third-party managed syslog", func() {

		Context("with the legacy syslog plugin", func() {

			Context("and tcp receiver", func() {

				BeforeEach(func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolTCP, false, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					const conf = `
<store>
	@type syslog_buffered
	@id syslogid
	remote_syslog syslog-receiver.openshift-logging.svc
	port 24224
	hostname ${hostname}
	facility user
	severity debug
</store>
					`
					//create configmap syslog/"syslog.conf"
					if err = e2e.CreateLegacySyslogConfigMap(syslogDeployment.Namespace, conf); err != nil {
						Fail(fmt.Sprintf("Unable to create legacy syslog.conf configmap: %v", err))
					}

					components := []helpers.LogComponentType{helpers.ComponentTypeCollector}
					cr := helpers.NewClusterLogging(components...)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
				})

				It("should send logs to the forward.Output logstore", func() {
					name := syslogDeployment.GetName()
					Expect(e2e.LogStores[name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				})
			})

			Context("and udp receiver", func() {

				BeforeEach(func() {
					if syslogDeployment, err = e2e.DeploySyslogReceiver(testDir, corev1.ProtocolUDP, false, helpers.RFC3164); err != nil {
						Fail(fmt.Sprintf("Unable to deploy syslog receiver: %v", err))
					}
					const conf = `
<store>
	@type syslog
	@id syslogid
	remote_syslog syslog-receiver.openshift-logging.svc
	port 24224
	hostname ${hostname}
	facility user
	severity debug
</store>
					`
					//create configmap syslog/"syslog.conf"
					if err = e2e.CreateLegacySyslogConfigMap(syslogDeployment.Namespace, conf); err != nil {
						Fail(fmt.Sprintf("Unable to create legacy syslog.conf configmap: %v", err))
					}

					components := []helpers.LogComponentType{helpers.ComponentTypeCollector, helpers.ComponentTypeStore}
					cr := helpers.NewClusterLogging(components...)
					if err := e2e.CreateClusterLogging(cr); err != nil {
						Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
					}
					for _, component := range components {
						if err := e2e.WaitFor(component); err != nil {
							Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
						}
					}
				})

				It("should send logs to the forward.Output logstore", func() {
					name := syslogDeployment.GetName()
					Expect(e2e.LogStores[name].HasInfraStructureLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored infrastructure logs")
				})
			})

		})

		AfterEach(func() {
			e2e.Cleanup()
			e2e.WaitForCleanupCompletion(helpers.OpenshiftLoggingNS, []string{"fluentd", "syslog-receiver", "elasticsearch"})
		})

	})

})
