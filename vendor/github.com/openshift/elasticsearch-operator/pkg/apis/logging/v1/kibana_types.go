package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KibanaSpec defines the desired state of Kibana
// +k8s:openapi-gen=true
type KibanaSpec struct {
	ManagementState ManagementState          `json:"managementState"`
	Resources       *v1.ResourceRequirements `json:"resources"`
	NodeSelector    map[string]string        `json:"nodeSelector,omitempty"`
	Tolerations     []v1.Toleration          `json:"tolerations,omitempty"`
	Replicas        int32                    `json:"replicas"`
	ProxySpec       `json:"proxy,omitempty"`
}

type ProxySpec struct {
	Resources *v1.ResourceRequirements `json:"resources"`
}

// KibanaStatus defines the observed state of Kibana
// +k8s:openapi-gen=true
type KibanaStatus struct {
	Replicas    int32                        `json:"replicas"`
	Deployment  string                       `json:"deployment"`
	ReplicaSets []string                     `json:"replicaSets"`
	Pods        PodStateMap                  `json:"pods"`
	Conditions  map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Kibana is the Schema for the kibanas API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kibanas,scope=Namespaced
type Kibana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KibanaSpec     `json:"spec,omitempty"`
	Status []KibanaStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KibanaList contains a list of Kibana
type KibanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kibana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kibana{}, &KibanaList{})
}
