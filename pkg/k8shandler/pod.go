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
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewPodSpec is a constructor to instaniate a new PodSpec.
// Notice that all Aggregated Logging relevant pods are (force-)allocated to linux nodes, see https://jira.coreos.com/browse/LOG-411
func NewPodSpec(serviceAccountName string, containers []core.Container, volumes []core.Volume, nodeSelector map[string]string, tolerations []core.Toleration) core.PodSpec {
	return core.PodSpec{
		Containers:         containers,
		ServiceAccountName: serviceAccountName,
		Volumes:            volumes,
		NodeSelector:       utils.EnsureLinuxNodeSelector(nodeSelector),
		Tolerations:        tolerations,
	}
}

//NewContainer stubs an instance of a Container
func NewContainer(containerName string, imageName string, pullPolicy core.PullPolicy, resources core.ResourceRequirements) core.Container {
	return core.Container{
		Name:            containerName,
		Image:           utils.GetComponentImage(imageName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}

//GetPodList for a given selector and namespace
func (clusterRequest *ClusterLoggingRequest) GetPodList(selector map[string]string) (*core.PodList, error) {
	list := &core.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: core.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(
		selector,
		list,
	)

	return list, err
}
