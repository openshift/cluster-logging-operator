package v1alpha1

import (
  "k8s.io/api/core/v1"
  "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterLoggingList struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ListMeta   `json:"metadata"`
  Items             []ClusterLogging `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterLogging struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata"`
  Spec              ClusterLoggingSpec   `json:"spec"`
  Status            ClusterLoggingStatus `json:"status,omitempty"`
}

type ClusterLoggingSpec struct {
  Visualization     VisualizationSpec `json:"visualization,omitempty"`
  LogStore          LogStoreSpec `json:"logStore,omitempty"`
  Collection        CollectionSpec `json:"collection,omitempty"`
  Curation          CurationSpec `json:"curation,omitempty"`
}

// This is the struct that will contain information pertinent to Log visualization (Kibana)
type VisualizationSpec struct {
  Type              string `json:"type"`
  KibanaSpec `json:"kibana,omitempty"`
}

type KibanaSpec struct {
  Resources         v1.ResourceRequirements `json:"resources"`
  NodeSelector      map[string]string `json:"nodeSelector,omitempty"`
  Replicas          int32 `json:"replicas"`
  ProxySpec `json:"proxy,omitempty"`
}

type ProxySpec struct {
  Resources         v1.ResourceRequirements `json:"resources"`
}

// This is the struct that will contain information pertinent to Log storage (Elasticsearch)
type LogStoreSpec struct {
  Type              string `json:"type"`
  ElasticsearchSpec `json:"elasticsearch,omitempty"`
}

type ElasticsearchSpec struct {
  Resources         v1.ResourceRequirements `json:"resources"`
  Replicas          int32 `json:"replicas"`
  NodeSelector      map[string]string `json:"nodeSelector,omitempty"`
  Storage           v1alpha1.ElasticsearchNodeStorageSource `json:"storage"`
}

// This is the struct that will contain information pertinent to Log collection (Fluentd)
type CollectionSpec struct {
  Type              string `json:"type"`
  FluentdSpec `json:"fluentd,omitempty"`
  Normalizer        NormalizerSpec `json:normalizerSpec,omitempty"`
}

type FluentdSpec struct {
  Resources         v1.ResourceRequirements `json:"resources"`
  NodeSelector      map[string]string `json:"nodeSelector,omitempty"`
}

// This is the struct that will contain information pertinent to Log normalization (Mux)
type NormalizerSpec struct {
  Type              string `json:"type"`
  FluentdSpec `json:"fluentd,omitempty"`
}

// This is the struct that will contain information pertinent to Log curation (Curator)
type CurationSpec struct {
  Type              string `json:"type"`
  CuratorSpec `json:"curator,omitempty"`
}

type CuratorSpec struct {
  Resources         v1.ResourceRequirements `json:"resources"`
  NodeSelector      map[string]string `json:"nodeSelector,omitempty"`
  Schedule          string `json:"schedule"`
}

type ClusterLoggingStatus struct {
  Visualization     VisualizationStatus `json:"visualization"`
  LogStore          LogStoreStatus `json:"logStore"`
  Collection        CollectionStatus `json:"collection"`
  Curation          CurationStatus `json:"curation"`
}

type VisualizationStatus struct {
  Type              string `json:"type"`
  KibanaStatus `json:",inline,omitempty"`
}

type KibanaStatus struct {
  Replicas          int32 `json:"replicas"`
  ReplicaSets       []string `json:"replicaSets"`
  Pods              []string `json:"pods"`
}

type LogStoreStatus struct {
  Type              string `json:"type"`
  ElasticsearchStatus `json:",inline,omitempty"`
}

type ElasticsearchStatus struct {
  Replicas          int32 `json:"replicas"`
  ReplicaSets       []string `json:"replicaSets"`
  Pods              []string `json:"pods"`
}

type CollectionStatus struct {
  Type              string `json:"type"`
  FluentdCollectorStatus `json:",inline,omitempty"`
  NormalizerStatus  NormalizerStatus `json:"normalizerStatus,omitempty"`
}

type FluentdCollectorStatus struct {
  DaemonSets        []string `json:"daemonSets"`
  Nodes             []string `json:"nodes"`
  Pods              []string `json:"pods"`
}

type FluentdNormalizerStatus struct {
  Replicas          int32 `json:"replicas"`
  ReplicaSets       []string `json:"replicaSets"`
  Pods              []string `json:"pods"`
}

type NormalizerStatus struct {
  Type              string `json:"type"`
  FluentdNormalizerStatus `json:",inline,omitempty"`
}

type CurationStatus struct {
  Type              string `json:"type"`
  CuratorStatus `json:",inline,omitempty"`
}

type CuratorStatus struct {
  ChronJobs         []string `json:"chronJobs"`
  Schedules         []string `json:"schedules"`
}
