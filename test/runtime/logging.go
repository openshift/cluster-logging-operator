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

package runtime

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
)

// NewClusterLogForwarder returns a ClusterLogForwarder with default name and namespace.
func NewClusterLogForwarder() *loggingv1.ClusterLogForwarder {
	clf := &loggingv1.ClusterLogForwarder{}
	Initialize(clf, test.OpenshiftLoggingNS, test.InstanceName)
	return clf
}

// NewClusterLogging returns a ClusterLogging with default name, namespace and
// collection configuration. No store, visualization or curation are configured,
// see ClusterLoggingDefaultXXX to add them.
func NewClusterLogging() *loggingv1.ClusterLogging {
	cl := &loggingv1.ClusterLogging{}
	Initialize(cl, test.OpenshiftLoggingNS, test.InstanceName)
	test.MustUnmarshal(`
    collection:
      logs:
        fluentd: {}
        type: fluentd
    managementState: Managed
    `, &cl.Spec)
	return cl
}

// ClusterLoggingDefaultStore sets default store configuration.
func ClusterLoggingDefaultStore(cl *loggingv1.ClusterLogging) {
	test.MustUnmarshal(`
    type: "elasticsearch"
    elasticsearch:
      nodeCount: 1
      redundancyPolicy: "ZeroRedundancy"
      resources:
        limits:
          cpu: 500m
          memory: 4Gi
`, &cl.Spec.LogStore)
}

// ClusterLoggingDefaultVisualization sets default visualization configuration.
func ClusterLoggingDefaultVisualization(cl *loggingv1.ClusterLogging) {
	test.MustUnmarshal(`
    type: "kibana"
    kibana:
      replicas: 1
`, &cl.Spec.Visualization)
}

// ClusterLoggingDefaultCuration sets defautl curation configuration.
func ClusterLoggingDefaultCuration(cl *loggingv1.ClusterLogging) {
	test.MustUnmarshal(`
    type: "curator"
    curator:
      schedule: "30 3,9,15,21 * * *"
`, &cl.Spec.Curation)
}
