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

package helpers

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clolog "github.com/ViaQ/logerr/log"
	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

type LogComponentType string

const (
	ComponentTypeStore         LogComponentType = "LogStore"
	ComponentTypeVisualization LogComponentType = "Visualization"
	ComponentTypeCollector     LogComponentType = "Collector"
)

func NewClusterLogging(componentTypes ...LogComponentType) *cl.ClusterLogging {
	clolog.Info("NewClusterLogging ", "componentTypes", componentTypes)
	instance := &cl.ClusterLogging{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: cl.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterLoggingName,
			Namespace: OpenshiftLoggingNS,
		},
		Spec: cl.ClusterLoggingSpec{
			ManagementState: cl.ManagementStateManaged,
		},
	}

	for _, compType := range componentTypes {
		switch compType {
		case ComponentTypeStore:
			instance.Spec.LogStore = &cl.LogStoreSpec{
				Type: cl.LogStoreTypeElasticsearch,
				ElasticsearchSpec: cl.ElasticsearchSpec{
					Resources: &v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2Gi"),
							v1.ResourceCPU:    resource.MustParse("100m"),
						},
					},
					ProxySpec: cl.ProxySpec{
						Resources: &v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("512Mi"),
								v1.ResourceCPU:    resource.MustParse("200m"),
							},
						},
					},
					NodeCount:        1,
					RedundancyPolicy: elasticsearch.ZeroRedundancy,
				},
			}
		case ComponentTypeCollector:
			instance.Spec.Collection = &cl.CollectionSpec{
				Logs: cl.LogCollectionSpec{
					Type: cl.LogCollectionTypeFluentd,
					FluentdSpec: cl.FluentdSpec{
						Resources: &v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("500Mi"),
								v1.ResourceCPU:    resource.MustParse("200m"),
							},
						},
					},
				},
			}
		}
	}
	return instance
}
