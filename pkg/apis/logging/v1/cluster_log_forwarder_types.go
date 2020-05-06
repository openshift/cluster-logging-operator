package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

const ClusterLogForwarderKind = "ClusterLogForwarder"

// ClusterLogForwarder is the schema for the `clusterlogforwarder` API.
//
// A forwarder defines:
// - `inputs` that select logs to be forwarded `outputs`
// - `outputs` that identify targets to receive logs.
// - `pipelines` that forward logs from a set of inputs to a set of outputs.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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

	// Pipelines forward messages from a set of inputs to a set of outputs.
	//
	// Pipelines refer to inputs and outputs by name. As well as then named inputs
	// and outputs defined in the ClusterLogForwarder resource, there are some
	// built-in names, see `inputRefs` and `outputRefs` for more.
	//
	// +required
	Pipelines []PipelineSpec `json:"pipelines,omitempty"`
}

type ClusterLogForwarderStatus struct {
	// Conditions of the log forwarder.
	Conditions status.Conditions `json:"conditions,omitempty"`
	// Inputs maps input names to conditions of the input.
	Inputs NamedConditions `json:"inputs,omitempty"`
	// Outputs maps output names to conditions of the output.
	Outputs NamedConditions `json:"outputs,omitempty"`
	// Pipelines maps pipeline names to conditions of the pipeline.
	Pipelines NamedConditions `json:"pipelines,omitempty"`
}

// IsReady returns true if all of the subordinate conditions are ready.
func (status ClusterLogForwarderStatus) IsReady() bool {
	for _, nc := range []NamedConditions{status.Pipelines, status.Inputs, status.Outputs} {
		for _, conds := range nc {
			if !conds.IsTrueFor(ConditionReady) {
				return false
			}
		}
	}
	return true
}

// IsDegraded returns true if any of the subordinate conditions are degraded.
func (status ClusterLogForwarderStatus) IsDegraded() bool {
	for _, nc := range []NamedConditions{status.Pipelines, status.Inputs, status.Outputs} {
		for _, conds := range nc {
			if conds.IsTrueFor(ConditionDegraded) {
				return true
			}
		}
	}
	return false
}

type PipelineSpec struct {
	// OutputRefs lists the names of outputs from this pipeline.
	//
	// The following built-in names are always available:
	//
	// - 'default' Output to the default log store provided by ClusterLogging.
	//
	// +required
	OutputRefs []string `json:"outputRefs"`

	// InputRefs lists the names of inputs to this pipeline.
	//
	// The following built-in names are always available:
	//
	// - 'application' Container application logs (excludes infrastructure containers)
	// - 'infrastructure' Infrastructure container logs and OS system logs.
	// - 'audit' System audit logs.
	//
	// +required
	InputRefs []string `json:"inputRefs"`

	// Name is optional, but must be unique in the `pipelines` list if provided.
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

// InputMap returns a map of names to outputs.
func (spec *ClusterLogForwarderSpec) InputMap() map[string]*InputSpec {
	m := map[string]*InputSpec{}
	for i := range spec.Inputs {
		m[spec.Inputs[i].Name] = &spec.Inputs[i]
	}
	return m
}
