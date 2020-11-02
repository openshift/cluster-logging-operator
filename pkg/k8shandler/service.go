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

package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/pkg/factory"
	core "k8s.io/api/core/v1"
)

//RemoveService with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveService(serviceName string) error {

	service := factory.NewService(
		serviceName,
		clusterRequest.Cluster.Namespace,
		serviceName,
		[]core.ServicePort{},
	)

	err := clusterRequest.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}
