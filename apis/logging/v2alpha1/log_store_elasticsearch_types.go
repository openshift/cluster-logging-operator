/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2alpha1

import (
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
)

type ElasticsearchSpec struct {
	// The resource requirements for Elasticsearch
	//
	// +nullable
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Elasticsearch Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Number of nodes to deploy for Elasticsearch
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Elasticsearch Size",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podCount"}
	NodeCount int32 `json:"nodeCount,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Elasticsearch Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:nodeSelector"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The storage specification for Elasticsearch data nodes
	//
	// +nullable
	// +optional
	Storage elasticsearch.ElasticsearchStorageSpec `json:"storage,omitempty"`

	// +optional
	RedundancyPolicy elasticsearch.RedundancyPolicyType `json:"redundancyPolicy,omitempty"`

	// Specification of the Elasticsearch Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

type RetentionPoliciesSpec struct {
	// +nullable
	App *RetentionPolicySpec `json:"application,omitempty"`

	// +nullable
	Infra *RetentionPolicySpec `json:"infra,omitempty"`

	// +nullable
	Audit *RetentionPolicySpec `json:"audit,omitempty"`
}

type RetentionPolicySpec struct {
	// +optional
	MaxAge elasticsearch.TimeUnit `json:"maxAge"`

	// How often to run a new prune-namespaces job
	// +optional
	PruneNamespacesInterval elasticsearch.TimeUnit `json:"pruneNamespacesInterval"`

	// The per namespace specification to delete documents older than a given minimum age
	// +optional
	Namespaces []elasticsearch.IndexManagementDeleteNamespaceSpec `json:"namespaceSpec,omitempty"`

	// The threshold percentage of ES disk usage that when reached, old indices should be deleted (e.g. 75)
	// +optional
	DiskThresholdPercent int64 `json:"diskThresholdPercent,omitempty"`
}

type ProxySpec struct {
	// +nullable
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ElasticsearchRoleType string

const (
	ElasticsearchRoleTypeClient ElasticsearchRoleType = "client"
	ElasticsearchRoleTypeData   ElasticsearchRoleType = "data"
	ElasticsearchRoleTypeMaster ElasticsearchRoleType = "master"
)

type ElasticsearchStatus struct {
	// +optional
	ClusterName string `json:"clusterName"`
	// +optional
	NodeCount int32 `json:"nodeCount"`
	// +optional
	ReplicaSets []string `json:"replicaSets,omitempty"`
	// +optional
	Deployments []string `json:"deployments,omitempty"`
	// +optional
	StatefulSets []string `json:"statefulSets,omitempty"`
	// +optional
	ClusterHealth string `json:"clusterHealth,omitempty"`
	// +optional
	Cluster elasticsearch.ClusterHealth `json:"cluster"`
	// +optional
	Pods map[ElasticsearchRoleType]PodStateMap `json:"pods,omitempty"`
	// +optional
	ShardAllocationEnabled elasticsearch.ShardAllocationState `json:"shardAllocationEnabled"`
	// +optional
	ClusterConditions []Conditions `json:"clusterConditions,omitempty"`
	// +optional
	NodeConditions map[string][]Conditions `json:"nodeConditions,omitempty"`
}
