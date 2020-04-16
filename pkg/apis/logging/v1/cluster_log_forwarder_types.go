package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

const ClusterLogForwarderKind = "ClusterLogForwarder"

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
	// Inputs are named inputs of log messages
	//
	// +required
	Inputs []InputSpec `json:"inputs,omitempty"`
	// Outputs are named destinations for log messages.
	//
	// +required
	Outputs []OutputSpec `json:"outputs,omitempty"`

	// Pipelines select log messages to send to outputs.
	//
	// +required
	Pipelines []PipelineSpec `json:"pipelines,omitempty"`
}

type ClusterLogForwarderStatus struct {
	// Conditions of the log forwarder.
	Conditions Conditions `json:"conditions,omitempty"`
	// Inputs maps input names to conditions of the input.
	Inputs NamedConditions `json:"inputs,omitempty"`
	// Outputs maps output names to conditions of the output.
	Outputs NamedConditions `json:"outputs,omitempty"`
	// Pipelines maps pipeline names to conditions of the pipeline.
	Pipelines NamedConditions `json:"pipelines,omitempty"`
}

type PipelineSpec struct {
	// OutputRefs lists the names of outputs from this pipeline.
	//
	// +required
	OutputRefs []string `json:"outputRefs"`

	// InputRefs lists the names of inputs to this pipeline.
	//
	// +required
	InputRefs []string `json:"inputRefs"`

	// Name is optional, but must be unique in the `pipelines` list if provided.
	// Required to allow patch updates to the pipelines list.
	//
	// +optional
	Name string `json:"name,omitempty"`
}

// ClusterLogForwarderList is a list of ClusterLogForwarders
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterLogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogForwarder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogForwarder{}, &ClusterLogForwarderList{})
}

// RouteMap maps input names to connected outputs or vice-versa.
type RouteMap map[string]sets.String

func (m RouteMap) Insert(k, v string) {
	if m[k] == nil {
		m[k] = sets.NewString()
	}
	m[k].Insert(v)
}

// Routes maps connected input and output names.
type Routes struct {
	ByInput, ByOutput RouteMap
}

func NewRoutes(pipelines []PipelineSpec) Routes {
	r := Routes{
		ByInput:  map[string]sets.String{},
		ByOutput: map[string]sets.String{},
	}
	for _, p := range pipelines {
		for _, inRef := range p.InputRefs {
			for _, outRef := range p.OutputRefs {
				r.ByInput.Insert(inRef, outRef)
				r.ByOutput.Insert(outRef, inRef)
			}
		}
	}
	return r
}

// OutputMap returns a map of names to outputs.
func (spec *ClusterLogForwarderSpec) OutputMap() map[string]*OutputSpec {
	m := map[string]*OutputSpec{}
	for i := range spec.Outputs {
		m[spec.Outputs[i].Name] = &spec.Outputs[i]
	}
	return m
}
