package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ClusterLogForwarderKind = "ClusterLogForwarder"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=clf

// ClusterLogForwarder is an API to configure forwarding logs.
//
// You configure forwarding by specifying a list of `pipelines`,
// which forward from a set of named inputs to a set of named outputs.
//
// There are built-in input names for common log categories, and you can
// define custom inputs to do additional filtering.
//
// There is a built-in output name for the default openshift log store, but
// you can define your own outputs with a URL and other connection information
// to forward logs to other stores or processors, inside or outside the cluster.
//
// For more details see the documentation on the API fields.
//
type ClusterLogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLogForwarderSpec   `json:"spec,omitempty"`
	Status ClusterLogForwarderStatus `json:"status,omitempty"`
}

// ClusterLogForwarderSpec defines the desired state of ClusterLogForwarder
type ClusterLogForwarderSpec struct {
	// Inputs are named filters for log messages to be forwarded.
	//
	// There are three built-in inputs named `application`, `infrastructure` and
	// `audit`. You don't need to define inputs here if those are sufficient for
	// your needs. See `inputRefs` for more.
	//
	// +optional
	Inputs []InputSpec `json:"inputs,omitempty"`

	// Outputs are named destinations for log messages.
	//
	// There is a built-in output named `default` which forwards to the default
	// openshift log store. You can define outputs to forward to other stores or
	// log processors, inside or outside the cluster.
	//
	// +optional
	Outputs []OutputSpec `json:"outputs,omitempty"`

	// Pipelines forward the messages selected by a set of inputs to a set of outputs.
	//
	// +required
	Pipelines []PipelineSpec `json:"pipelines,omitempty"`

	// OutputDefaults are used to specify default values for OutputSpec
	//
	// +optional
	OutputDefaults *OutputDefaults `json:"outputDefaults,omitempty"`
}

// ClusterLogForwarder represents the current status of ClusterLogForwarder
type ClusterLogForwarderStatus struct {
	// Conditions of the log forwarder.
	Conditions status.Conditions `json:"conditions,omitempty"`
	// Inputs maps input name to condition of the input.
	Inputs NamedConditions `json:"inputs,omitempty"`
	// Outputs maps output name to condition of the output.
	Outputs NamedConditions `json:"outputs,omitempty"`
	// Pipelines maps pipeline name to condition of the pipeline.
	Pipelines NamedConditions `json:"pipelines,omitempty"`
}

// InputSpec defines a selector of log messages.
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +kubebuilder:validation:minLength:=1
	// +required
	Name string `json:"name"`

	// NOTE: the following fields in this struct are deliberately _not_ `omitempty`.
	// An empty field means enable that input type with no filter.

	// Application, if present, enables `application` logs.
	//
	// +optional
	Application *Application `json:"application,omitempty"`

	// Infrastructure, if present, enables `infrastructure` logs.
	//
	// +optional
	Infrastructure *Infrastructure `json:"infrastructure,omitempty"`

	// Audit, if present, enables `audit` logs.
	//
	// +optional
	Audit *Audit `json:"audit,omitempty"`
}

type Application struct {
	// Namespaces is a list of namespaces from which to collect application logs.
	// If the list is empty, logs are collected from all namespaces.
	//
	// +optional
	Namespaces []string `json:"namespaces"`
	// Selector selects logs from all pods with matching labels.
	// For testing purpose, MatchLabels is only supported.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// Infrastructure enables infrastructure logs. Filtering may be added in future.
type Infrastructure struct{}

// Audit enables audit logs. Filtering may be added in future.
type Audit struct{}

// Output defines a destination for log messages.
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +kubebuilder:validation:minLength:=1
	// +required
	Name string `json:"name"`

	// Type of output plugin.
	//
	// +kubebuilder:validation:Enum:=syslog;fluentdForward;elasticsearch;kafka;cloudwatch;loki
	// +required
	Type string `json:"type"`

	// URL to send log records to.
	//
	// An absolute URL, with a scheme. Valid schemes depend on `type`.
	// Special schemes `tcp`, `tls`, `udp` and `udps` are used for types that
	// have no scheme of their own. For example, to send syslog records using secure UDP:
	//
	//     { type: syslog, url: udps://syslog.example.com:1234 }
	//
	// Basic TLS is enabled if the URL scheme requires it (for example 'https' or 'tls').
	// The 'username@password' part of `url` is ignored.
	// Any additional authentication material is in the `secret`.
	// See the `secret` field for more details.
	//
	// +kubebuilder:validation:Pattern:=`^$|[a-zA-z]+:\/\/.*`
	// +optional
	URL string `json:"url,omitempty"`

	OutputTypeSpec `json:",inline"`

	// Secret for authentication.
	// Name of a secret in the same namespace as the cluster logging operator.
	//
	// For client authentication, set secret keys `tls.crt` and `tls.key` to the client certificate and private key.
	//
	// To use your own certificate authority, set secret key `ca-bundle.crt`.
	//
	// Depending on the `type` there may be other secret keys that have meaning.
	//
	// +optional
	Secret *OutputSecretSpec `json:"secret,omitempty"`
}

// OutputSecretSpec is a secret reference containing name only, no namespace.
type OutputSecretSpec struct {
	// Name of a secret in the namespace configured for log forwarder secrets.
	//
	// +required
	Name string `json:"name"`
}

type PipelineSpec struct {
	// OutputRefs lists the names (`output.name`) of outputs from this pipeline.
	//
	// The following built-in names are always available:
	//
	// 'default' Output to the default log store provided by ClusterLogging.
	//
	// +required
	OutputRefs []string `json:"outputRefs"`

	// InputRefs lists the names (`input.name`) of inputs to this pipeline.
	//
	// The following built-in input names are always available:
	//
	// `application` selects all logs from application pods.
	//
	// `infrastructure` selects logs from openshift and kubernetes pods and some node logs.
	//
	// `audit` selects node logs related to security audits.
	//
	// +required
	InputRefs []string `json:"inputRefs"`

	// Labels lists labels applied to this pipeline
	//
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Name is optional, but must be unique in the `pipelines` list if provided.
	//
	// +optional
	Name string `json:"name,omitempty"`

	// Parse enables parsing of log entries into structured logs
	//
	// Logs are parsed according to parse value, only `json` is supported as of now.
	//
	// +kubebuilder:validation:Enum:=json
	// +optional
	Parse string `json:"parse,omitempty"`
}

type OutputDefaults struct {

	// Elasticsearch OutputSpec default values
	//
	// Values specified here will be used as default values for Elasticsearch Output spec
	//
	// +optional
	Elasticsearch *Elasticsearch `json:"elasticsearch,omitempty"`
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
