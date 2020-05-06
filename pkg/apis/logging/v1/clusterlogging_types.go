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
	ManagementState ManagementState    `json:"managementState"`
	Visualization   *VisualizationSpec `json:"visualization,omitempty"`
	LogStore        *LogStoreSpec      `json:"logStore,omitempty"`
	Collection      *CollectionSpec    `json:"collection,omitempty"`
	Curation        *CurationSpec      `json:"curation,omitempty"`
}

// ClusterLoggingStatus defines the observed state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Visualization VisualizationStatus `json:"visualization"`
	LogStore      LogStoreStatus      `json:"logStore"`
	Collection    CollectionStatus    `json:"collection"`
	Curation      CurationStatus      `json:"curation"`
	Conditions    status.Conditions   `json:"clusterConditions,omitempty"`
}

// This is the struct that will contain information pertinent to Log visualization (Kibana)
type VisualizationSpec struct {
	Type       VisualizationType `json:"type"`
	KibanaSpec `json:"kibana,omitempty"`
}

type KibanaSpec struct {
	Resources    *v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration          `json:"tolerations,omitempty"`
	Replicas     int32                    `json:"replicas"`
	ProxySpec    `json:"proxy,omitempty"`
}

type ProxySpec struct {
	Resources *v1.ResourceRequirements `json:"resources"`
}

// This is the struct that will contain information pertinent to Log storage (Elasticsearch)
type LogStoreSpec struct {
	Type              LogStoreType `json:"type"`
	ElasticsearchSpec `json:"elasticsearch,omitempty"`
	RetentionPolicy   *RetentionPoliciesSpec `json:"retentionPolicy,omitempty"`
}

type RetentionPoliciesSpec struct {
	App   *RetentionPolicySpec `json:"application,omitempty"`
	Infra *RetentionPolicySpec `json:"infra,omitempty"`
	Audit *RetentionPolicySpec `json:"audit,omitempty"`
}

type RetentionPolicySpec struct {
	MaxAge elasticsearch.TimeUnit `json:"maxAge"`
}

type ElasticsearchSpec struct {
	Resources        *v1.ResourceRequirements               `json:"resources"`
	NodeCount        int32                                  `json:"nodeCount"`
	NodeSelector     map[string]string                      `json:"nodeSelector,omitempty"`
	Tolerations      []v1.Toleration                        `json:"tolerations,omitempty"`
	Storage          elasticsearch.ElasticsearchStorageSpec `json:"storage"`
	RedundancyPolicy elasticsearch.RedundancyPolicyType     `json:"redundancyPolicy"`
}

// This is the struct that will contain information pertinent to Log and event collection
type CollectionSpec struct {
	Logs LogCollectionSpec `json:"logs,omitempty"`
}

type LogCollectionSpec struct {
	Type        LogCollectionType `json:"type"`
	FluentdSpec `json:"fluentd,omitempty"`
}

type EventCollectionSpec struct {
	Type EventCollectionType `json:"type"`
}

type FluentdSpec struct {
	Resources    *v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration          `json:"tolerations,omitempty"`
}

// This is the struct that will contain information pertinent to Log curation (Curator)
type CurationSpec struct {
	Type        CurationType `json:"type"`
	CuratorSpec `json:"curator,omitempty"`
}

type CuratorSpec struct {
	Resources    *v1.ResourceRequirements `json:"resources"`
	NodeSelector map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration          `json:"tolerations,omitempty"`
	Schedule     string                   `json:"schedule"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClusterLogging is the Schema for the clusterloggings API
// +k8s:openapi-gen=true
type ClusterLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLoggingSpec   `json:"spec,omitempty"`
	Status ClusterLoggingStatus `json:"status,omitempty"`
}

type VisualizationStatus struct {
	KibanaStatus []elasticsearch.KibanaStatus `json:"kibanaStatus,omitempty"`
}

type KibanaStatus struct {
	Replicas    int32                        `json:"replicas"`
	Deployment  string                       `json:"deployment"`
	ReplicaSets []string                     `json:"replicaSets"`
	Pods        PodStateMap                  `json:"pods"`
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
