package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

const (
	ServiceAccountName string = "elasticsearch"
	ConfigMapName      string = "elasticsearch"
	SecretName         string = "elasticsearch"
)

// ElasticsearchList struct represents list of Elasticsearch objects
type ElasticsearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Elasticsearch `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Elasticsearch struct represents Elasticsearch cluster CRD
type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ElasticsearchSpec   `json:"spec"`
	Status            ElasticsearchStatus `json:"status,omitempty"`
}

type ReplicationPolicyType string

const (
	FullReplication    ReplicationPolicyType = "FullReplication"
	PartialReplication ReplicationPolicyType = "PartialReplication"
	NoReplication      ReplicationPolicyType = "NoReplication"
)

// ElasticsearchSpec struct represents the Spec of Elasticsearch cluster CRD
type ElasticsearchSpec struct {
	// managementState indicates whether and how the operator should manage the component
	ManagementState   ManagementState       `json:"managementState"`
	ReplicationPolicy ReplicationPolicyType `json:"dataReplication"`
	Nodes             []ElasticsearchNode   `json:"nodes"`
	Spec              ElasticsearchNodeSpec `json:"nodeSpec"`
}

// ElasticsearchNode struct represents individual node in Elasticsearch cluster
type ElasticsearchNode struct {
	Roles        []ElasticsearchNodeRole        `json:"roles"`
	Replicas     int32                          `json:"replicas"`
	Spec         ElasticsearchNodeSpec          `json:"nodeSpec"`
	NodeSelector map[string]string              `json:"nodeSelector,omitempty"`
	Storage      ElasticsearchNodeStorageSource `json:"storage"`
}

type ElasticsearchNodeStorageSource struct {
	// HostPath option will mount directory from the host.
	// Cluster administrator must grant `hostaccess` scc to the service account.
	// Cluster admin also must set appropriate SELINUX labels and perissions
	// for the directory on the host.
	HostPath *v1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// EmptyDir should be only used for testing purposes and not in production.
	// This option will use temporary directory for data storage. Data will be lost
	// when Pod is regenerated.
	EmptyDir *v1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// VolumeClaimTemplate is supposed to act similarly to VolumeClaimTemplates field
	// of StatefulSetSpec. Meaning that it'll generate a number of PersistentVolumeClaims
	// per individual Elasticsearch cluster node. The actual PVC name used will
	// be constructed from VolumeClaimTemplate name, node type and replica number
	// for the specific node.
	VolumeClaimTemplate *v1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// PersistentVolumeClaim will NOT try to regenerate PVC, it will be used
	// as-is. You may want to use it instead of VolumeClaimTemplate in case
	// you already have bounded PersistentVolumeClaims you want to use, and the names
	// of these PersistentVolumeClaims doesn't follow the naming convention.
	PersistentVolumeClaim *v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty"`
}

// ElasticsearchNodeStatus represents the status of individual Elasticsearch node
type ElasticsearchNodeStatus struct {
	DeploymentName  string                  `json:"deploymentName,omitempty"`
	ReplicaSetName  string                  `json:"replicaSetName,omitempty"`
	StatefulSetName string                  `json:"statefulSetName,omitempty"`
	PodName         string                  `json:"podName,omitempty"`
	Status          string                  `json:"status,omitempty"`
	Roles           []ElasticsearchNodeRole `json:"roles,omitempty"`
}

// ElasticsearchNodeSpec represents configuration of an individual Elasticsearch node
type ElasticsearchNodeSpec struct {
	Image     string                  `json:"image,omitempty"`
	Resources v1.ResourceRequirements `json:"resources"`
}

type ElasticsearchRequiredAction string

const (
	ElasticsearchActionRollingRestartNeeded ElasticsearchRequiredAction = "RollingRestartNeeded"
	ElasticsearchActionFullRestartNeeded    ElasticsearchRequiredAction = "FullRestartNeeded"
	ElasticsearchActionInterventionNeeded   ElasticsearchRequiredAction = "InterventionNeeded"
	ElasticsearchActionNewClusterNeeded     ElasticsearchRequiredAction = "NewClusterNeeded"
	ElasticsearchActionNone                 ElasticsearchRequiredAction = "ClusterOK"
	ElasticsearchActionScaleDownNeeded      ElasticsearchRequiredAction = "ScaleDownNeeded"
)

type ElasticsearchNodeRole string

const (
	ElasticsearchRoleClient ElasticsearchNodeRole = "client"
	ElasticsearchRoleData   ElasticsearchNodeRole = "data"
	ElasticsearchRoleMaster ElasticsearchNodeRole = "master"
)

// ElasticsearchStatus represents the status of Elasticsearch cluster
type ElasticsearchStatus struct {
	Nodes         []ElasticsearchNodeStatus             `json:"nodes"`
	ClusterHealth string                                `json:"clusterHealth"`
	Pods          map[ElasticsearchNodeRole]PodStateMap `json:"pods"`
	Conditions    []ClusterCondition                    `json:"conditions"`
}

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)

type ManagementState string

const (
	// Managed means that the operator is actively managing its resources and trying to keep the component active.
	// It will only upgrade the component if it is safe to do so
	ManagementStateManaged ManagementState = "Managed"
	// Unmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

// ClusterCondition contains details for the current condition of this elasticsearch cluster.
type ClusterCondition struct {
	// Type is the type of the condition.
	Type ClusterConditionType `json:"type"`
	// Status is the status of the condition.
	Status ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// ClusterConditionType is a valid value for ClusterCondition.Type
type ClusterConditionType string

// These are valid conditions for elasticsearch node
const (
	UpdatingSettings ClusterConditionType = "UpdatingSettings"
	ScalingUp        ClusterConditionType = "ScalingUp"
	ScalingDown      ClusterConditionType = "ScalingDown"
	Restarting       ClusterConditionType = "Restarting"
)

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type ClusterEvent string

const (
	ScaledDown            ClusterEvent = "ScaledDown"
	ScaledUp              ClusterEvent = "ScaledUp"
	UpdateClusterSettings ClusterEvent = "UpdateClusterSettings"
	NoEvent               ClusterEvent = "NoEvent"
)
