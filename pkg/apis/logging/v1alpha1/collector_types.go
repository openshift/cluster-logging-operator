package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Collector is an instance of a collector
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Collector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CollectorSpec `json:"spec,omitempty"`
}

//CollectorSpec is the specification for deployable collectors
//
// +k8s:openapi-gen=true
type CollectorSpec struct {
	// The type of Log Collection to configure
	//
	// +kubebuilder:validation:Enum:=promtail
	Type CollectorType `json:"spec,omitempty"`

	// +nullable
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	PromTailSpec `json:"promtail,omitempty"`
}

type PromTailSpec struct {
	Endpoint string `json:"endpoint,omitempty"`
}

//CollectorType is a kind of collector
type CollectorType string

const (
	CollectorKind string = "Collector"

	CollectorTypePromtail CollectorType = "promtail"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CollectorList contains a list of Collector
type CollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CollectorSpec `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Collector{}, &CollectorList{})
}
