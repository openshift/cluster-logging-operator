package v1

import (
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLoggingSpec defines the desired state of ClusterLogging
type ClusterLoggingSpec struct {
	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator
	ManagementState ManagementState `json:"managementState,omitempty"`
	// Specification of the Visualization component for the cluster
	Visualization *VisualizationSpec `json:"visualization,omitempty"`
	// Specification of the Log Storage component for the cluster
	LogStore *LogStoreSpec `json:"logStore,omitempty"`
	// Specification of the Collection component for the cluster
	Collection *CollectionSpec `json:"collection,omitempty"`
	// Specification of the Curation component for the cluster
	Curation *CurationSpec `json:"curation,omitempty"`
}

// ClusterLoggingStatus defines the observed state of ClusterLogging
type ClusterLoggingStatus struct {
	Visualization VisualizationStatus `json:"visualization"`
	LogStore      LogStoreStatus      `json:"logStore"`
	Collection    CollectionStatus    `json:"collection"`
	Curation      CurationStatus      `json:"curation"`
	Conditions    []ClusterCondition  `json:"clusterConditions,omitempty"`
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
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`
	// Number of instances to deploy for a Kibana deployment
	Replicas int32 `json:"replicas"`
	// Specification of the Kibana Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

type ProxySpec struct {
	// The resource requirements for Kibana
	Resources *v1.ResourceRequirements `json:"resources"`
}

// This is the struct that will contain information pertinent to Log storage (Elasticsearch)
type LogStoreSpec struct {
	// The type of Log Storage to configure
	Type LogStoreType `json:"type"`
	// Specification of the Elasticsearch Log Store component
	ElasticsearchSpec `json:"elasticsearch,omitempty"`
	// Retention policy defines the maximum age for an index after which it should be deleted
	RetentionPolicy *RetentionPoliciesSpec `json:"retentionPolicy,omitempty"`
}

type RetentionPoliciesSpec struct {
	App   *RetentionPolicySpec `json:"application,omitempty"`
	Infra *RetentionPolicySpec `json:"infra,omitempty"`
	Audit *RetentionPolicySpec `json:"audit,omitempty"`
}

type RetentionPolicySpec struct {
	// Maximum age as integer followed by time unit; one of [yMwdhHms]
	MaxAge elasticsearch.TimeUnit `json:"maxAge,omitempty"`
}

type ElasticsearchSpec struct {
	// The resource requirements for Kibana
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// Number of nodes to deploy for Elasticsearch
	NodeCount int32 `json:"nodeCount"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`
	// The storage specification for Elasticsearch data nodes
	Storage          elasticsearch.ElasticsearchStorageSpec `json:"storage,omitempty"`
	RedundancyPolicy elasticsearch.RedundancyPolicyType     `json:"redundancyPolicy,omitempty"`
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
	Resources *v1.ResourceRequirements `json:"resources"`
	// Define which Nodes the Pods are scheduled on.
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
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`
	// The cron schedule that the Curator job is run. Defaults to "30 3 * * *"
	Schedule string `json:"schedule"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type ClusterLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Logging cluster.
	Spec ClusterLoggingSpec `json:"spec,omitempty"`
	// Observed state of ClusterLogging
	Status ClusterLoggingStatus `json:"status,omitempty"`
}

type VisualizationStatus struct {
	KibanaStatus []elasticsearch.KibanaStatus `json:"kibanaStatus,omitempty"`
}

type KibanaStatus struct {
	// Number of instances to deploy for a Kibana deployment
	Replicas    int32                        `json:"replicas,omitempty"`
	Deployment  string                       `json:"deployment,omitempty"`
	ReplicaSets []string                     `json:"replicaSets,omitempty"`
	Pods        PodStateMap                  `json:"pods,omitempty"`
	Conditions  map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type LogStoreStatus struct {
	ElasticsearchStatus []ElasticsearchStatus `json:"elasticsearchStatus,omitempty"`
}

type ElasticsearchStatus struct {
	ClusterName            string                                    `json:"clusterName"`
	NodeCount              int32                                     `json:"nodeCount"`
	ReplicaSets            []string                                  `json:"replicaSets,omitempty"`
	Deployments            []string                                  `json:"deployments,omitempty"`
	StatefulSets           []string                                  `json:"statefulSets,omitempty"`
	ClusterHealth          string                                    `json:"clusterHealth,omitempty"`
	Cluster                elasticsearch.ClusterHealth               `json:"cluster"`
	Pods                   map[ElasticsearchRoleType]PodStateMap     `json:"pods"`
	ShardAllocationEnabled elasticsearch.ShardAllocationState        `json:"shardAllocationEnabled"`
	ClusterConditions      ElasticsearchClusterConditions            `json:"clusterConditions,omitempty"`
	NodeConditions         map[string]ElasticsearchClusterConditions `json:"nodeConditions,omitempty"`
}

type CollectionStatus struct {
	Logs LogCollectionStatus `json:"logs,omitempty"`
}

type LogCollectionStatus struct {
	FluentdStatus FluentdCollectorStatus `json:"fluentdStatus,omitempty"`
}

type EventCollectionStatus struct {
}

type FluentdCollectorStatus struct {
	DaemonSet  string                       `json:"daemonSet"`
	Nodes      map[string]string            `json:"nodes"`
	Pods       PodStateMap                  `json:"pods"`
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type FluentdNormalizerStatus struct {
	Replicas    int32                        `json:"replicas"`
	ReplicaSets []string                     `json:"replicaSets"`
	Pods        PodStateMap                  `json:"pods"`
	Conditions  map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type NormalizerStatus struct {
	FluentdStatus []FluentdNormalizerStatus `json:"fluentdStatus,omitempty"`
}

type CurationStatus struct {
	CuratorStatus []CuratorStatus `json:"curatorStatus,omitempty"`
}

type CuratorStatus struct {
	CronJob    string                       `json:"cronJobs"`
	Schedule   string                       `json:"schedules"`
	Suspended  bool                         `json:"suspended"`
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

// +kubebuilder:validation:Enum=Managed;Unmanaged
type ManagementState string

const (
	// Managed means that the operator is actively managing its resources and trying to keep the component active.
	// It will only upgrade the component if it is safe to do so
	ManagementStateManaged ManagementState = "Managed"
	// Unmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

type ClusterCondition struct {
	Type   ClusterConditionType `json:"type"`
	Status v1.ConditionStatus   `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// +kubebuilder:validation:Enum=IncorrectCRName;ContainerWaiting;ContainerTerminated;Unschedulable;NodeStorage
type ClusterConditionType string

const (
	IncorrectCRName     ClusterConditionType = "IncorrectCRName"
	ContainerWaiting    ClusterConditionType = "ContainerWaiting"
	ContainerTerminated ClusterConditionType = "ContainerTerminated"
	Unschedulable       ClusterConditionType = "Unschedulable"
	NodeStorage         ClusterConditionType = "NodeStorage"
	CollectorDeadEnd    ClusterConditionType = "CollectorDeadEnd"
)

// `operator-sdk generate crds` does not allow map-of-slice, must use named types.

type ClusterConditions []ClusterCondition
type ElasticsearchClusterConditions []elasticsearch.ClusterCondition

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClusterLoggingList contains a list of ClusterLogging
type ClusterLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogging{}, &ClusterLoggingList{})
}
