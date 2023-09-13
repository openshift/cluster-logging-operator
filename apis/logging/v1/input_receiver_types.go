package v1

// NOTE: The Enum validation on ReceiverSpec.Type must be updated if the list of types changes.

// ReceiverSpec is a union of input Receiver types.
//
// The fields of this struct define the set of known Receiver types.
type ReceiverSpec struct {
	HTTP *HTTPReceiver `json:"http,omitempty"`
}

const (
	FormatK8SAudit = "k8s_audit" // Log events in k8s list format, e.g. API audit log events.
)

// HTTPReceiver receives encoded logs as a HTTP endpoint.
type HTTPReceiver struct {
	// Port that this receiver will listen on.
	// Used to create a Service for this port.
	//
	// +required
	Port int32 `json:"port"`
	// Format is the format of incoming log data.
	//
	// +kubebuilder:validation:Enum:=k8s_audit
	// +required
	Format string `json:"format"`
}
