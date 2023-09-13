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

// ReceiverPort specifies parameters for the Service fronting the HTTPReceiver
type ReceiverPort struct {
	// Name of the service to create for this HTTPReceiver
	// If not specified, defaults to the name of the containing ClusterLogForwarder input
	// +optional
	Name string `json:"name,omitempty"`

	// Port the Service will listen on.
	//
	// +required
	Port int32 `json:"port"`

	// Port the Receiver will listen on.
	// If not specified, defaults to the value of Port
	// +optional
	TargetPort int32 `json:"targetPort,omitempty"`
}

// HTTPReceiver receives encoded logs as a HTTP endpoint.
type HTTPReceiver struct {
	// ReceiverPort specifies parameters for the Service fronting the HTTPReceiver
	// +required
	ReceiverPort ReceiverPort `json:"receiverPort"`
	// Format is the format of incoming log data.
	//
	// +kubebuilder:validation:Enum:=kubeAPIAudit
	// +required
	Format string `json:"format"`
}
