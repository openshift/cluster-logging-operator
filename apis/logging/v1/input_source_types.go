package v1

import (
	corev1 "k8s.io/api/core/v1"
)

// NOTE: The Enum validation on SourceSpec.Type must be updated if the list of types changes.

// SourceSpec is a union of input source types.
//
// The fields of this struct define the set of known source types.
type SourceSpec struct {
	HTTP *HTTPSource `json:"http,omitempty"`
}

const (
	FormatK8S  = "k8s"  // Log events in k8s list format, e.g. API audit log events.
	FormatText = "text" // Newline-separated plain-text log records.
)

// HTTPSource receives encoded logs as a HTTP endpoint.
type HTTPSource struct {

	// FIXME TLS server configuration - can this be handled in the Service,
	// or do we need vector to have certificates? Can the service be bypassed?

	// FIXME What to do about service name clashes? Can't generate random name, service name is used to locate the source.

	// Name of the source.
	// A Service resource is generated for each source,
	// the name must be a unique Service name in the ClusterLogForwarder's namespace.
	//
	// +required
	Name string `json:"name"`

	// Port that this source will listen on.
	// Used to create a Service for this port.
	//
	// +required
	Port corev1.ServicePort `json:"port"`
	// Format is the format of incoming log data.
	//
	// +kubebuilder:validation:Enum:=k8s;text
	// +required
	Format string `json:"format"`

	// LogType indicates the type of logs received from this source.
	//
	// +kubebuilder:validation:Enum:=application;infrastructure;audit
	// +required
	LogType string `json:"logType"`
}
