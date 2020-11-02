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
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

var _ = Describe("[Collection] Fluentd message filtering", func() {
	_, filename, _, _ := runtime.Caller(0)
	log.Info("Running ", "filename", filename)
	var (
		err              error
		fluentDeployment *apps.Deployment
		e2e              = helpers.NewE2ETestFramework()
		rootDir          string
	)

	BeforeEach(func() {
		if err := e2e.DeployLogGenerator(); err != nil {
			Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
		}
		rootDir = filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "/")
		if fluentDeployment, err = e2e.DeployFluentdReceiver(rootDir, false); err != nil {
			Fail(fmt.Sprintf("Unable to deploy fluent receiver: %v", err))
		}

		forwarder := &logging.ClusterLogForwarder{
			TypeMeta: metav1.TypeMeta{
				Kind:       logging.ClusterLogForwarderKind,
				APIVersion: logging.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance",
			},
			Spec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Name: fluentDeployment.ObjectMeta.Name,
						Type: logging.OutputTypeFluentdForward,
						URL:  fmt.Sprintf("tcp://%s.%s.svc:24224", fluentDeployment.ObjectMeta.Name, fluentDeployment.Namespace),
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "test-app",
						OutputRefs: []string{fluentDeployment.ObjectMeta.Name},
						InputRefs:  []string{logging.InputNameApplication},
					},
				},
			},
		}
		if err := e2e.CreateClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of clusterlogforwarder: %v", err))
		}
		cr := helpers.NewClusterLogging(helpers.ComponentTypeCollector)
		if err := e2e.CreateClusterLogging(cr); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of cluster logging: %v", err))
		}
		if err := e2e.WaitFor(helpers.ComponentTypeCollector); err != nil {
			Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", helpers.ComponentTypeCollector, err))
		}

	})

	AfterEach(func() {
		e2e.Cleanup()
	}, helpers.DefaultCleanUpTimeout)

	It("should remove 'kubernetes.labels' and create 'kubernetes.flat_labels' with an array of 'kubernetes.labels'", func() {
		Expect(e2e.LogStores[fluentDeployment.GetName()].HasApplicationLogs(helpers.DefaultWaitForLogsTimeout)).To(BeTrue(), "Expected to find stored application logs")

		//verify infra namespaces are not stored to their own index
		logs, err := e2e.LogStores[fluentDeployment.GetName()].ApplicationLogs(helpers.DefaultWaitForLogsTimeout)
		Expect(err).To(BeNil(), fmt.Sprintf("Error fetching logs: %v", err))
		Expect(len(logs)).To(Not(Equal(0)), "There were no documents returned in the logs")

		//verify the new key exists
		Expect(logs[0].Kubernetes.FlatLabels).To(Not(BeNil()), fmt.Sprintf("Expected to find the kubernetes.flat_labels key in %#v", logs[0]))

		//verify we removed the old key
		Expect(logs[0].Kubernetes.Labels).To(BeNil(), fmt.Sprintf("Did not expect to find the kubernetes.labels key in %#v", logs[0]))
	})

})
