package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/inputs"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

// Built-in log input names
const (
	InputApplication    = "Application"    // Containers from non-infrastructure namespaces
	InputInfrastructure = "Infrastructure" // Infrastructure containers and system logs
	InputAudit          = "Audit"          // System audit logs
)

var BuiltInInputs = sets.NewString(InputApplication, InputInfrastructure, InputAudit)

// InputSpec defines a source of log messages.
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// Type of input source.
	// Must be one of :
	//  - "application"
	//  - "infrastructure"
	//  - "audit"
	//
	// +kubebuilder:validation:Enum:=application;infrastructure;audit
	// +required
	Type string `json:"type"`

	// InputTypeSpec is inlined with a required `type` and optional extra configuration.
	//
	// +optional
	InputTypeSpec `json:",inline,omitempty"`
}

// InputTypeSpec is a union of optional type-specific extra specs.
//
// This is the single source of truth for relating input type names to spec
// classes.
type InputTypeSpec struct {
	// Application specific specs
	// +optional
	Application *inputs.ApplicationType `json:"application,omitempty"`
}
