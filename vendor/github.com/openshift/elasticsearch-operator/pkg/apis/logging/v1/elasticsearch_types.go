package v1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServiceAccountName string = "elasticsearch"
	ConfigMapName      string = "elasticsearch"
	SecretName         string = "elasticsearch"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Elasticsearch is the Schema for the elasticsearches API
// +k8s:openapi-gen=true
type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchSpec   `json:"spec,omitempty"`
	Status ElasticsearchStatus `json:"status,omitempty"`
}

//AddOwnerRefTo appends the Elasticsearch object as an OwnerReference to the passed object
func (es *Elasticsearch) AddOwnerRefTo(o metav1.Object) {
	trueVar := true
	ref := metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       "Elasticsearch",
		Name:       es.Name,
		UID:        es.UID,
		Controller: &trueVar,
	}
	if (metav1.OwnerReference{}) != ref {
		o.SetOwnerReferences(append(o.GetOwnerReferences(), ref))
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchList contains a list of Elasticsearch
type ElasticsearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Elasticsearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Elasticsearch{}, &ElasticsearchList{})
}

// ElasticsearchSpec defines the desired state of Elasticsearch
// +k8s:openapi-gen=true
// ManagementState indicates whether and how the operator should manage the component
type ElasticsearchSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	ManagementState  ManagementState       `json:"managementState"`
	RedundancyPolicy RedundancyPolicyType  `json:"redundancyPolicy"`
	Nodes            []ElasticsearchNode   `json:"nodes"`
	Spec             ElasticsearchNodeSpec `json:"nodeSpec"`
	IndexManagement  *IndexManagementSpec  `json:"indexManagement"`
}

// ElasticsearchStatus defines the observed state of Elasticsearch
// +k8s:openapi-gen=true
type ElasticsearchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	Nodes                  []ElasticsearchNodeStatus             `json:"nodes"`
	ClusterHealth          string                                `json:"clusterHealth"`
	Cluster                ClusterHealth                         `json:"cluster"`
	ShardAllocationEnabled ShardAllocationState                  `json:"shardAllocationEnabled"`
	Pods                   map[ElasticsearchNodeRole]PodStateMap `json:"pods"`
	Conditions             ClusterConditions                     `json:"conditions"`
	IndexManagementStatus  *IndexManagementStatus                `json:"indexManagement,omitempty"`
}

type ClusterHealth struct {
	Status              string `json:"status"`
	NumNodes            int32  `json:"numNodes"`
	NumDataNodes        int32  `json:"numDataNodes"`
	ActivePrimaryShards int32  `json:"activePrimaryShards"`
	ActiveShards        int32  `json:"activeShards"`
	RelocatingShards    int32  `json:"relocatingShards"`
	InitializingShards  int32  `json:"initializingShards"`
	UnassignedShards    int32  `json:"unassignedShards"`
	PendingTasks        int32  `json:"pendingTasks"`
}

// ElasticsearchNode struct represents individual node in Elasticsearch cluster
// GenUUID will be populated by the operator if not provided
type ElasticsearchNode struct {
	Roles          []ElasticsearchNodeRole  `json:"roles"`
	NodeCount      int32                    `json:"nodeCount"`
	Resources      v1.ResourceRequirements  `json:"resources"`
	NodeSelector   map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations    []v1.Toleration          `json:"tolerations,omitempty"`
	Storage        ElasticsearchStorageSpec `json:"storage"`
	GenUUID        *string                  `json:"genUUID,omitempty"`
	ProxyResources v1.ResourceRequirements  `json:"proxyResources,omitempty"`
}

// ElasticsearchNodeSpec represents configuration of an individual Elasticsearch node
type ElasticsearchNodeSpec struct {
	Image          string                  `json:"image,omitempty"`
	Resources      v1.ResourceRequirements `json:"resources"`
	NodeSelector   map[string]string       `json:"nodeSelector,omitempty"`
	Tolerations    []v1.Toleration         `json:"tolerations,omitempty"`
	ProxyResources v1.ResourceRequirements `json:"proxyResources,omitempty"`
}

type ElasticsearchStorageSpec struct {
	// The class of storage to provision. More info: https://kubernetes.io/docs/concepts/storage/storage-classes/
	StorageClassName *string `json:"storageClassName,omitempty"`
	// The capacity of storage to provision.
	Size *resource.Quantity `json:"size,omitempty"`
}

// ElasticsearchNodeStatus represents the status of individual Elasticsearch node
type ElasticsearchNodeStatus struct {
	DeploymentName  string                         `json:"deploymentName,omitempty"`
	StatefulSetName string                         `json:"statefulSetName,omitempty"`
	Status          string                         `json:"status,omitempty"`
	UpgradeStatus   ElasticsearchNodeUpgradeStatus `json:"upgradeStatus,omitempty"`
	Roles           []ElasticsearchNodeRole        `json:"roles,omitempty"`
	Conditions      ClusterConditions              `json:"conditions,omitempty"`
}

type ElasticsearchNodeUpgradeStatus struct {
	ScheduledForUpgrade      v1.ConditionStatus        `json:"scheduledUpgrade,omitempty"`
	ScheduledForRedeploy     v1.ConditionStatus        `json:"scheduledRedeploy,omitempty"`
	ScheduledForCertRedeploy v1.ConditionStatus        `json:"scheduledCertRedeploy,omitempty"`
	UnderUpgrade             v1.ConditionStatus        `json:"underUpgrade,omitempty"`
	UpgradePhase             ElasticsearchUpgradePhase `json:"upgradePhase,omitempty"`
}

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

type ClusterConditions []ClusterCondition

// +kubebuilder:validation:Enum=FullRedundancy;MultipleRedundancy;SingleRedundancy;ZeroRedundancy

// The policy towards data redundancy to specify the number of redundant primary shards
type RedundancyPolicyType string

const (
	// FullRedundancy - each index is fully replicated on every Data node in the cluster
	FullRedundancy RedundancyPolicyType = "FullRedundancy"
	// MultipleRedundancy - each index is spread over half of the Data nodes
	MultipleRedundancy RedundancyPolicyType = "MultipleRedundancy"
	// SingleRedundancy - one replica shard
	SingleRedundancy RedundancyPolicyType = "SingleRedundancy"
	// ZeroRedundancy - no replica shards
	ZeroRedundancy RedundancyPolicyType = "ZeroRedundancy"
)

type ElasticsearchNodeRole string

const (
	ElasticsearchRoleClient ElasticsearchNodeRole = "client"
	ElasticsearchRoleData   ElasticsearchNodeRole = "data"
	ElasticsearchRoleMaster ElasticsearchNodeRole = "master"
)

type ShardAllocationState string

const (
	ShardAllocationAll       ShardAllocationState = "all"
	ShardAllocationNone      ShardAllocationState = "none"
	ShardAllocationPrimaries ShardAllocationState = "primaries"
	ShardAllocationUnknown   ShardAllocationState = "shard allocation unknown"
)

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)

type ElasticsearchUpgradePhase string

const (
	NodeRestarting      ElasticsearchUpgradePhase = "nodeRestarting"
	RecoveringData      ElasticsearchUpgradePhase = "recoveringData"
	ControllerUpdated   ElasticsearchUpgradePhase = "controllerUpdated"
	PreparationComplete ElasticsearchUpgradePhase = "preparationComplete"
)

// Managed means that the operator is actively managing its resources and trying to keep the component active.
// It will only upgrade the component if it is safe to do so
// Unmanaged means that the operator will not take any action related to the component
type ManagementState string

const (
	ManagementStateManaged   ManagementState = "Managed"
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

// ClusterConditionType is a valid value for ClusterCondition.Type
type ClusterConditionType string

const (
	UpdatingSettings         ClusterConditionType = "UpdatingSettings"
	ScalingUp                ClusterConditionType = "ScalingUp"
	ScalingDown              ClusterConditionType = "ScalingDown"
	Restarting               ClusterConditionType = "Restarting"
	Recovering               ClusterConditionType = "Recovering"
	UpdatingESSettings       ClusterConditionType = "UpdatingESSettings"
	InvalidMasters           ClusterConditionType = "InvalidMasters"
	InvalidData              ClusterConditionType = "InvalidData"
	InvalidRedundancy        ClusterConditionType = "InvalidRedundancy"
	InvalidUUID              ClusterConditionType = "InvalidUUID"
	ESContainerWaiting       ClusterConditionType = "ElasticsearchContainerWaiting"
	ESContainerTerminated    ClusterConditionType = "ElasticsearchContainerTerminated"
	ProxyContainerWaiting    ClusterConditionType = "ProxyContainerWaiting"
	ProxyContainerTerminated ClusterConditionType = "ProxyContainerTerminated"
	Unschedulable            ClusterConditionType = "Unschedulable"
	NodeStorage              ClusterConditionType = "NodeStorage"
	CustomImage              ClusterConditionType = "CustomImageIgnored"
)
