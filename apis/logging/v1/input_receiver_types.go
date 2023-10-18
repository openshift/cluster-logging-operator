package v1

// NOTE: The Enum validation on ReceiverSpec.Type must be updated if the list of types changes.

// ReceiverSpec is a union of input Receiver types.
//
// The fields of this struct define the set of known Receiver types.
type ReceiverSpec struct {
	HTTP *HTTPReceiver `json:"http,omitempty"`
}

const (
	FormatKubeAPIAudit = "kubeAPIAudit" // Log events in k8s list format, e.g. API audit log events.
)

// HTTPReceiver receives encoded logs as a HTTP endpoint.
type HTTPReceiver struct {
	// Port the Service and the HTTP listener listen on.
	// +kubebuilder:default:=8443
	// +optional
	Port int32 `json:"port"`

	// Format is the format of incoming log data.
	//
	// +kubebuilder:validation:Enum:=kubeAPIAudit
	// +required
	Format string `json:"format"`
}
