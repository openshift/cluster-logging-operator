package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/status"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLoggingSpec defines the desired state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator
	//
	// +kubebuilder:validation:Enum:=Managed;Unmanaged
	// +optional
	ManagementState ManagementState `json:"managementState"`

	// Specification of the Visualization component for the cluster
	//
	// +nullable
	Visualization *VisualizationSpec `json:"visualization,omitempty"`

	// Specification of the Log Storage component for the cluster
	//
	// +nullable
	LogStore *LogStoreSpec `json:"logStore,omitempty"`

	// Specification of the Collection component for the cluster
	//
	// +nullable
	Collection *CollectionSpec `json:"collection,omitempty"`

	// Specification of the Curation component for the cluster
	//
	// +nullable
	Curation *CurationSpec `json:"curation,omitempty"`
}

// ClusterLoggingStatus defines the observed state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +optional
	Visualization VisualizationStatus `json:"visualization"`

	// +optional
	LogStore LogStoreStatus `json:"logStore"`

	// +optional
	Collection CollectionStatus `json:"collection"`

	// +optional
	Curation CurationStatus `json:"curation"`

	// +optional
	Conditions status.Conditions `json:"clusterConditions,omitempty"`
}

// This is the struct that will contain information pertinent to Log visualization (Kibana)
type VisualizationSpec struct {
	// The type of Visualization to configure
	Type VisualizationType `json:"type"`

	// Specification of the Kibana Visualization component
	KibanaSpec `json:"kibana,omitempty"`
}

type KibanaSpec struct {
	// The resource requirements for Kibana
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// Number of instances to deploy for a Kibana deployment
	Replicas int32 `json:"replicas"`

	// Specification of the Kibana Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

type ProxySpec struct {
	// +nullable
	Resources *v1.ResourceRequirements `json:"resources"`
}

// This is the struct that will contain information pertinent to Log storage (Elasticsearch)
type LogStoreSpec struct {
	// The type of Log Storage to configure
	Type LogStoreType `json:"type"`

	// Specification of the Elasticsearch Log Store component
	ElasticsearchSpec `json:"elasticsearch,omitempty"`

	// Retention policy defines the maximum age for an index after which it should be deleted
	//
	// +nullable
	RetentionPolicy *RetentionPoliciesSpec `json:"retentionPolicy,omitempty"`
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
}

type ElasticsearchSpec struct {
	// The resource requirements for Elasticsearch
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Number of nodes to deploy for Elasticsearch
	NodeCount int32 `json:"nodeCount"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The storage specification for Elasticsearch data nodes
	//
	// +nullable
	// +optional
	Storage elasticsearch.ElasticsearchStorageSpec `json:"storage"`

	// +optional
	RedundancyPolicy elasticsearch.RedundancyPolicyType `json:"redundancyPolicy"`

	// Specification of the Elasticsearch Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

// This is the struct that will contain information pertinent to Log and event collection
type CollectionSpec struct {
	// Specification of Log Collection for the cluster
	Logs LogCollectionSpec `json:"logs,omitempty"`
}

type LogCollectionSpec struct {
	// The type of Log Collection to configure
	Type LogCollectionType `json:"type"`

	// Specification of the Fluentd Log Collection component
	FluentdSpec `json:"fluentd,omitempty"`
}

type EventCollectionSpec struct {
	Type EventCollectionType `json:"type"`
}

type FluentdSpec struct {
	// The resource requirements for Fluentd
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`
}

// This is the struct that will contain information pertinent to Log curation (Curator)
type CurationSpec struct {
	// The kind of curation to configure
	Type CurationType `json:"type"`

	// The specification of curation to configure
	CuratorSpec `json:"curator,omitempty"`
}

type CuratorSpec struct {
	// The resource requirements for Curator
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The cron schedule that the Curator job is run. Defaults to "30 3 * * *"
	Schedule string `json:"schedule"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClusterLogging is the Schema for the clusterloggings API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cl
//
// ClusterLogging is the Schema for the clusterloggings API
type ClusterLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Logging cluster.
	Spec ClusterLoggingSpec `json:"spec,omitempty"`

	// ClusterLoggingStatus defines the observed state of ClusterLogging
	Status ClusterLoggingStatus `json:"status,omitempty"`
}

type VisualizationStatus struct {
	// +optional
	KibanaStatus []elasticsearch.KibanaStatus `json:"kibanaStatus,omitempty"`
}

type KibanaStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Deployment string `json:"deployment"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type LogStoreStatus struct {
	// +optional
	ElasticsearchStatus []ElasticsearchStatus `json:"elasticsearchStatus,omitempty"`
}

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
	Pods map[ElasticsearchRoleType]PodStateMap `json:"pods"`
	// +optional
	ShardAllocationEnabled elasticsearch.ShardAllocationState `json:"shardAllocationEnabled"`
	// +optional
	ClusterConditions ElasticsearchClusterConditions `json:"clusterConditions,omitempty"`
	// +optional
	NodeConditions map[string]ElasticsearchClusterConditions `json:"nodeConditions,omitempty"`
}

type CollectionStatus struct {
	// +optional
	Logs LogCollectionStatus `json:"logs,omitempty"`
}

type LogCollectionStatus struct {
	// +optional
	FluentdStatus FluentdCollectorStatus `json:"fluentdStatus,omitempty"`
}

type EventCollectionStatus struct {
}

type FluentdCollectorStatus struct {
	// +optional
	DaemonSet string `json:"daemonSet"`
	// +optional
	Nodes map[string]string `json:"nodes"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type FluentdNormalizerStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type NormalizerStatus struct {
	// +optional
	FluentdStatus []FluentdNormalizerStatus `json:"fluentdStatus,omitempty"`
}

type CurationStatus struct {
	// +optional
	CuratorStatus []CuratorStatus `json:"curatorStatus,omitempty"`
}

type CuratorStatus struct {
	// +optional
	CronJob string `json:"cronJobs"`
	// +optional
	Schedule string `json:"schedules"`
	// +optional
	Suspended bool `json:"suspended"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)

type LogStoreType string

const (
	LogStoreTypeElasticsearch LogStoreType = "elasticsearch"
)

type ElasticsearchRoleType string

const (
	ElasticsearchRoleTypeClient ElasticsearchRoleType = "client"
	ElasticsearchRoleTypeData   ElasticsearchRoleType = "data"
	ElasticsearchRoleTypeMaster ElasticsearchRoleType = "master"
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

type ManagementState string

const (
	// Managed means that the operator is actively managing its resources and trying to keep the component active.
	// It will only upgrade the component if it is safe to do so
	ManagementStateManaged ManagementState = "Managed"
	// Unmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

const (
	IncorrectCRName     status.ConditionType = "IncorrectCRName"
	ContainerWaiting    status.ConditionType = "ContainerWaiting"
	ContainerTerminated status.ConditionType = "ContainerTerminated"
	Unschedulable       status.ConditionType = "Unschedulable"
	NodeStorage         status.ConditionType = "NodeStorage"
	CollectorDeadEnd    status.ConditionType = "CollectorDeadEnd"
)

// `operator-sdk generate crds` does not allow map-of-slice, must use a named type.
type ClusterConditions status.Conditions
type ElasticsearchClusterConditions []elasticsearch.ClusterCondition

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterLoggingList contains a list of ClusterLogging
type ClusterLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogging{}, &ClusterLoggingList{})
}
