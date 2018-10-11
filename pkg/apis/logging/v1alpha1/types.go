package v1alpha1

import (
	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterLogging `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterLoggingSpec   `json:"spec"`
	Status            ClusterLoggingStatus `json:"status,omitempty"`
}

type ClusterLoggingSpec struct {
	Visualization VisualizationSpec `json:"visualization,omitempty"`
	LogStore      LogStoreSpec      `json:"logStore,omitempty"`
	Collection    CollectionSpec    `json:"collection,omitempty"`
	Curation      CurationSpec      `json:"curation,omitempty"`
}

// This is the struct that will contain information pertinent to Log visualization (Kibana)
type VisualizationSpec struct {
	Type       VisualizationType `json:"type"`
	KibanaSpec `json:"kibana,omitempty"`
}

type KibanaSpec struct {
	Resources    v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string       `json:"nodeSelector,omitempty"`
	Replicas     int32                   `json:"replicas"`
	ProxySpec    `json:"proxy,omitempty"`
}

type ProxySpec struct {
	Resources v1.ResourceRequirements `json:"resources"`
}

// This is the struct that will contain information pertinent to Log storage (Elasticsearch)
type LogStoreSpec struct {
	Type              LogStoreType `json:"type"`
	ElasticsearchSpec `json:"elasticsearch,omitempty"`
}

type ElasticsearchSpec struct {
	Resources    v1.ResourceRequirements                 `json:"resources"`
	Replicas     int32                                   `json:"replicas"`
	NodeSelector map[string]string                       `json:"nodeSelector,omitempty"`
	Storage      v1alpha1.ElasticsearchNodeStorageSource `json:"storage"`
}

// This is the struct that will contain information pertinent to Log and event collection
type CollectionSpec struct {
	LogCollection   LogCollectionSpec   `json:"logCollection,omitempty"`
	EventCollection EventCollectionSpec `json:"eventCollection,omitempty"`
	Normalizer      NormalizerSpec      `json:"normalizer,omitempty"`
}

type LogCollectionSpec struct {
	Type        LogCollectionType `json:"type"`
	FluentdSpec `json:"fluentd,omitempty"`
}

type EventCollectionSpec struct {
	Type EventCollectionType `json:"type"`
}

type FluentdSpec struct {
	Resources    v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string       `json:"nodeSelector,omitempty"`
}

// This is the struct that will contain information pertinent to Log normalization (Mux)
type NormalizerSpec struct {
	Type        NormalizerType `json:"type"`
	FluentdSpec `json:"fluentd,omitempty"`
}

// This is the struct that will contain information pertinent to Log curation (Curator)
type CurationSpec struct {
	Type        CurationType `json:"type"`
	CuratorSpec `json:"curator,omitempty"`
}

type CuratorSpec struct {
	Resources    v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string       `json:"nodeSelector,omitempty"`
	Schedule     string                  `json:"schedule"`
}

type ClusterLoggingStatus struct {
	Visualization VisualizationStatus `json:"visualization"`
	LogStore      LogStoreStatus      `json:"logStore"`
	Collection    CollectionStatus    `json:"collection"`
	Curation      CurationStatus      `json:"curation"`
}

type VisualizationStatus struct {
	KibanaStatus []KibanaStatus `json:"kibanaStatus,omitempty"`
}

type KibanaStatus struct {
	Replicas    int32    `json:"replicas"`
	Deployment  string   `json:"deployment"`
	ReplicaSets []string `json:"replicaSets"`
	Pods        []string `json:"pods"`
}

type LogStoreStatus struct {
	ElasticsearchStatus []ElasticsearchStatus `json:"elasticsearchStatus,omitempty"`
}

type ElasticsearchStatus struct {
	ClusterName  string   `json:"clusterName"`
	Replicas     int32    `json:"replicas"`
	ReplicaSets  []string `json:"replicaSets"`
	Deployments  []string `json:"deployments"`
	StatefulSets []string `json:"statefulSets"`
	Pods         []string `json:"pods"`
}

type CollectionStatus struct {
	LogCollection    LogCollectionStatus   `json:"logCollection,omitempty"`
	EventCollection  EventCollectionStatus `json:"eventCollection,omitempty"`
	NormalizerStatus NormalizerStatus      `json:"normalizerStatus,omitempty"`
}

type LogCollectionStatus struct {
	FluentdStatus FluentdCollectorStatus `json:"fluentdStatus,omitempty"`
}

type EventCollectionStatus struct {
}

type FluentdCollectorStatus struct {
	DaemonSet string            `json:"daemonSet"`
	Pods      map[string]string `json:"pods"`
}

type FluentdNormalizerStatus struct {
	Replicas    int32    `json:"replicas"`
	ReplicaSets []string `json:"replicaSets"`
	Pods        []string `json:"pods"`
}

type NormalizerStatus struct {
	FluentdStatus []FluentdNormalizerStatus `json:"fluentdStatus,omitempty"`
}

type CurationStatus struct {
	CuratorStatus []CuratorStatus `json:"curatorStatus,omitempty"`
}

type CuratorStatus struct {
	CronJob   string `json:"cronJobs"`
	Schedule  string `json:"schedules"`
	Suspended bool   `json:"suspended"`
}

type LogStoreType string

const (
	LogStoreTypeElasticsearch LogStoreType = "elasticsearch"
)

type VisualizationType string

const (
	VisualizationTypeKibana VisualizationType = "kibana"
)

type CurationType string

const (
	CurationTypeCurator CurationType = "curator"
)

type LogCollectionType string

const (
	LogCollectionTypeFluentd LogCollectionType = "fluentd"
)

type EventCollectionType string

type NormalizerType string
