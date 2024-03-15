package v1

// NOTE: The Enum validation on ReceiverSpec.Type must be updated if the list of types changes.

// Receiver type constants, must match JSON tags of OutputTypeSpec fields.
const (
	ReceiverTypeHttp   = "http"
	ReceiverTypeSyslog = "syslog"

	FormatKubeAPIAudit = "kubeAPIAudit" // Log events in k8s list format, e.g. API audit log events.
)

// ReceiverSpec is a union of input Receiver types.
//
// The fields of this struct define the set of known Receiver types.
type ReceiverSpec struct {

	// Type of Receiver plugin.
	// +optional
	Type string `json:"type"`

	// The ReceiverTypeSpec that handles particular parameters
	// +optional
	*ReceiverTypeSpec `json:",inline,omitempty"`
}

type ReceiverTypeSpec struct {
	HTTP   *HTTPReceiver   `json:"http,omitempty"`
	Syslog *SyslogReceiver `json:"syslog,omitempty"`
}

// HTTPReceiver receives encoded logs as a HTTP endpoint.
type HTTPReceiver struct {
	// Port the Receiver listens on. It must be a value between 1024 and 65535
	// +kubebuilder:default:=8443
	// +kubebuilder:validation:Minimum:=1024
	// +kubebuilder:validation:Maximum:=65535
	// +optional
	Port int32 `json:"port"`

	// Format is the format of incoming log data.
	//
	// +kubebuilder:validation:Enum:=kubeAPIAudit
	// +required
	Format string `json:"format"`
}

// SyslogReceiver receives logs from rsyslog
type SyslogReceiver struct {
	// Port the Receiver listens on. It must be a value between 1024 and 65535
	// +kubebuilder:default:=10514
	// +kubebuilder:validation:Minimum:=1024
	// +kubebuilder:validation:Maximum:=65535
	// +optional
	Port int32 `json:"port"`
}
