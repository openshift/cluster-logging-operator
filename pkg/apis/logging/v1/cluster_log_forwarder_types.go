package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ClusterLogForwarder is the schema for the `clusterlogforwarder` API.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=cluster
// +kubebuilder:subresource:status
type ClusterLogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLogForwarderSpec   `json:"spec,omitempty"`
	Status ClusterLogForwarderStatus `json:"status,omitempty"`
}

// ClusterLogForwarderSpec defines the desired state of ClusterLogForwarder
type ClusterLogForwarderSpec struct {
	// Outputs are named destinations for log messages.
	//
	// +required
	Outputs []Output `json:"outputs,omitempty"`

	// Pipelines select log messages to send to outputs.
	//
	// +required
	Pipelines []Pipeline `json:"pipelines,omitempty"`
}

// NOTE: InputRef is initially restricted to built-in categories, in future it
// will be relaxed to allow reference to user-defined selector inputs.

// +kubebuilder:validation:Enum=Application;Infrastructure;Audit
type InputRef string

const (
	Application    InputRef = "Application"
	Infrastructure          = "Infrastructure"
	Audit                   = "Audit"
)

type Pipeline struct {
	// OutputNames lists the names of outputs from this pipeline.
	//
	// +required
	OutputRefs []string `json:"outputRefs"`

	// InputRefs lists the names of inputs to this pipeline.
	// By default, all available log streams are forwarded.
	//
	// +optional
	InputRefs []InputRef `json:"inputRefs,omitempty"`
}

// ClusterLogForwarderList is a list of ClusterLogForwarders
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterLogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogForwarder `json:"items"`
}

// FIXME(alanconway) review status - should have state/reason enums?

type ClusterLogForwarderStatus struct {
	State       string      `json:"state,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	Message     string      `json:"message,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
	Inputs      []SubStatus `json:"inputs,omitempty"`
	Outputs     []SubStatus `json:"outputs,omitempty"`
}

type SubStatus struct {
	//Name of the sub-object corresponding to for this status
	Name    string `json:"name,omitempty"`
	State   string `json:"state,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`

	Conditions  []Condition `json:"conditions,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

type Condition struct {
	Type        string      `json:"type,omitempty"`
	Status      string      `json:"status,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	Message     string      `json:"message,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogForwarder{}, &ClusterLogForwarderList{})
}
