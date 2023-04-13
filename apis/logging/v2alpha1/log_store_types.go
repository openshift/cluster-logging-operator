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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogStore identifies the default store to be used by a LogForwarder
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=ls
type LogStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogStoreSpec   `json:"spec,omitempty"`
	Status LogStoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type LogStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogStore `json:"items"`
}

// LogStoreSpec identifies the default store for logs.
type LogStoreSpec struct {
	// The Type of Log Storage to configure. The operator currently supports either using ElasticSearch
	// managed by elasticsearch-operator or Loki managed by loki-operator (LokiStack) as a default log store.
	//
	// When using ElasticSearch as a log store this operator also manages the ElasticSearch deployment.
	//
	// When using LokiStack as a log store this operator does not manage the LokiStack, but only creates
	// configuration referencing an existing LokiStack deployment. The user is responsible for creating and
	// managing the LokiStack himself.
	//
	// +kubebuilder:validation:Enum=elasticsearch;lokistack
	Type LogStoreType `json:"type"`

	// Specification of the Elasticsearch Log Store component
	Elasticsearch *ElasticsearchSpec `json:"elasticsearch,omitempty"`

	// LokiStack contains information about which LokiStack to use for log storage if Type is set to LogStoreTypeLokiStack.
	//
	// The cluster-logging-operator does not create or manage the referenced LokiStack.
	LokiStack *LokiStackStoreSpec `json:"lokistack,omitempty"`
}

// LokiStackStoreSpec is used to set up cluster-logging to use a LokiStack as logging storage.
// It points to an existing LokiStack in the same namespace.
type LokiStackStoreSpec struct {
	// Name of the LokiStack resource.
	//
	// +required
	Name string `json:"name"`
}

type LogStoreStatus struct {
	// +optional
	ElasticsearchStatus []ElasticsearchStatus `json:"elasticsearchStatus,omitempty"`
}

type LogStoreType string

const (
	//  NOTE: update the +kubebuilder:validation:Enum comment on LogStoreSpec.Type if you add values here.
	LogStoreTypeElasticsearch LogStoreType = "elasticsearch"
	LogStoreTypeLokiStack     LogStoreType = "lokistack"
)

type LogCollectionType string

const (
	LogCollectionTypeFluentd LogCollectionType = "fluentd"
	LogCollectionTypeVector  LogCollectionType = "vector"
)

func (ct LogCollectionType) IsSupportedCollector() bool {
	return ct == LogCollectionTypeFluentd || ct == LogCollectionTypeVector
}

type ManagementState string

const (
	// Managed means that the operator is actively managing its resources and trying to keep the component active.
	// It will only upgrade the component if it is safe to do so
	ManagementStateManaged ManagementState = "Managed"
	// Unmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

func init() {
	SchemeBuilder.Register(&LogStore{}, &LogStoreList{})
}
