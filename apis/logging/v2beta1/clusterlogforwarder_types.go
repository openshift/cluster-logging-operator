/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2beta1

import (
	"time"

	openshiftv1 "github.com/openshift/api/config/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ClusterLogForwarderKind = "ClusterLogForwarder"

// ClusterLogForwarderSpec specifies log forwarding for the entire cluster.
type ClusterLogForwarderSpec struct {

	// Inputs are named filters for log messages to be forwarded.
	//
	// There are three built-in inputs named `application`, `infrastructure` and
	// `audit`. You don't need to define inputs here if those are sufficient for
	// your needs. See `inputRefs` for more.
	//
	//+optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Inputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderInputs"}
	Inputs []InputSpec `json:"inputs,omitempty"`

	// Outputs are named destinations for log messages.
	//
	//+optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Outputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderOutputs"}
	Outputs []OutputSpec `json:"outputs,omitempty"`

	// Filters are named selection or transformation rules that can be applied to log records.
	//
	// There are different types of filter that can select and modify log records in different ways.
	// See [FilterTypeSpec] for a list of filter types.
	Filters []FilterSpec `json:"filters,omitempty"`

	// Pipelines forward the messages selected by a set of inputs to a set of outputs.
	//
	//+required
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Pipelines",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderPipelines"}
	Pipelines []PipelineSpec `json:"pipelines,omitempty"`

	// Namespace for resources associated with the ClusterLogForwarder.
	//
	// Resources named in this spec (for example the ServiceAccountName) must be in this namespace.
	// Resources created by the forwarder will be in this namespace.
	//
	//+required
	Namespace string `json:"namespace"`

	// ServiceAccountName is the name of a serviceaccount in the forwarders associated namespace.
	//
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// ClusterLogForwarderStatus defines the observed state of ClusterLogForwarder
type ClusterLogForwarderStatus struct {
	// Conditions of the log forwarder.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Forwarder Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderConditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Inputs maps input name to condition of the input.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Input Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:inputConditions"}
	Inputs map[string]metav1.Condition `json:"inputs,omitempty"`

	// Outputs maps output name to condition of the output.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Output Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:outputConditions"}
	Outputs map[string]metav1.Condition `json:"outputs,omitempty"`

	// Filters maps filter name to condition of the filter.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Filter Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:filterConditions"}
	Filters map[string]metav1.Condition `json:"filters,omitempty"`

	// Pipelines maps pipeline name to condition of the pipeline.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pipeline Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:pipelineConditions"}
	Pipelines map[string]metav1.Condition `json:"pipelines,omitempty"`
}

// InputSpec defines a selector of log messages for a given log type. The input is rejected
// if more than one of the following subfields are defined: application, infrastructure, audit, and receiver.
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +kubebuilder:validation:minLength:=1
	// +required
	Name string `json:"name"`

	// NOTE: the following fields in this struct are deliberately _not_ `omitempty`.
	// An empty field means enable that input type with no filter.

	// Application, if present, enables named set of `application` logs that
	// can specify a set of match criteria
	//
	// +optional
	Application *Application `json:"application,omitempty"`

	// Infrastructure, if present, enables `infrastructure` logs.
	//
	// +optional
	Infrastructure *Infrastructure `json:"infrastructure,omitempty"`

	// Audit, if present, enables `audit` logs.
	//
	//+optional
	Audit *Audit `json:"audit,omitempty"`

	// Receiver to receive logs from non-cluster sources.
	//
	//+optional
	Receiver *ReceiverSpec `json:"receiver,omitempty"`

	// Tuning parameters for the input.
	//
	//+optional
	Tuning *InputTuningSpec `json:"tuning,omitempty"`
}

// InputTuningSpec provides tuning parameters for an input.
type InputTuningSpec struct {
	// RateLimitPerContainer defines a rate limit on the maximum records-per-second
	// that each container can produce. Excess logs are dropped if the limit is exceeded.
	RateLimitPerContainer int `json:"rateLimitPerContainer"`
}

// Output defines a destination for log messages.
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +kubebuilder:validation:minLength:=1
	// +required
	Name string `json:"name"`

	// Type of output plugin.
	//
	// +kubebuilder:validation:Enum:=syslog;fluentdForward;elasticsearch;kafka;cloudwatch;loki;googleCloudLogging;splunk;http
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

	// TLS contains settings for controlling options on TLS client connections.
	TLS *OutputTLSSpec `json:"tls,omitempty"`

	// Secret for authentication.
	//
	// Names a secret in the same namespace as the ClusterLogForwarder.
	// Sensitive authentication information is stored in a separate Secret object.
	// A Secret is like a ConfigMap, where the keys are strings and the values are
	// base64-encoded binary data, for example TLS certificates.
	//
	// Common keys are described here.
	// Some output types support additional keys, documented with the output-specific configuration field.
	// All secret keys are optional, enable the security features you want by setting the relevant keys.
	//
	// Transport Layer Security (TLS)
	//
	// Using a TLS URL (`https://...` or `tls://...`) without any secret enables basic TLS:
	// client authenticates server using system default certificate authority.
	//
	// Additional TLS features are enabled by referencing a Secret with the following optional fields in its spec.data.
	// All data fields are base64 encoded.
	//
	//   * `tls.crt`: A client certificate, for mutual authentication. Requires `tls.key`.
	//   * `tls.key`: Private key to unlock the client certificate. Requires `tls.crt`
	//   * `passphrase`: Passphrase to decode an encoded TLS private key. Requires tls.key.
	//   * `ca-bundle.crt`: Custom CA to validate certificates.
	//
	// Username and Password
	//
	//   * `username`: Authentication user name. Requires `password`.
	//   * `password`: Authentication password. Requires `username`.
	//
	// Simple Authentication Security Layer (SASL)
	//
	//   * `sasl.enable`: (boolean) Explicitly enable or disable SASL.
	//     If missing, SASL is automatically enabled if any `sasl.*` keys are set.
	//   * `sasl.mechanisms`: (array of string) List of allowed SASL mechanism names.
	//     If missing or empty, the system defaults are used.
	//   * `sasl.allow-insecure`: (boolean) Allow mechanisms that send clear-text passwords.
	//     Default false.
	//
	// +optional
	Secret *OutputSecretSpec `json:"secret,omitempty"`

	// Tuning parameters for the output
	Tuning *OutputTuningSpec `json:"tuning,omitempty"`
}

// OutputTuningSpec tuning parameters for an output
type OutputTuningSpec struct {
	// RateLimit imposes a limit in records-per-second on the total aggregate rate of logs forwarded
	// by this output. Logs may be dropped to enforce the limit.
	// Missing or 0 means no rate limit.
	//
	// +optional
	RateLimit int `json:"rateLimit,omitempty"`

	// Delivery mode for log forwarding.
	//
	// - AtLeastOnce (default): if the forwarder crashes or is re-started, any logs that were read before
	//   the crash but not sent to their destination will be re-read and re-sent. Note it is possible
	//   that some logs are duplicated in the event of a crash - log recrods are deilvered at-least-once.
	// - AtMostOnce: The forwarder makes no effort to recover logs lost during a crash. This mode may give
	//   better throughput, but will result in more log loss.
	//
	// +required
	// +kubebuilder:validation:Enum:=AtLeastOnce;AtMostOnce
	Delivery string `json:"delivery,omitempty"`

	// Compression causes data to be compressed before sending on the network.
	// It is an error if the compression type is is not supported by the  output.
	Compression string `json:"compression,omitempty"`

	// MaxSendBytes limits the maximum payload of a single "send" to the output.
	// The default is set to an efficient value based on the output type.
	//
	// +optional
	MaxSendBytes resource.Quantity `json:"maxSendBytes,omitempty"`

	// MinRetryDuration is the minimum time to wait between attempts to re-connect after a failure.
	// The default is set to an efficient value based on the output type.
	//
	// +optional
	MinRetryDuration time.Duration `json:"minRetryWait,omitempty"`

	// MaxRetryDuration is the maximum time to wait between re-connect attempts after a connection failure.
	// The default is set to an efficient value based on the output type.
	//
	// +optional
	MaxRetryDuration time.Duration `json:"maxRetryWait,omitempty"`
}

// OutputTLSSpec contains options for TLS connections that are agnostic to the output type.
type OutputTLSSpec struct {
	// If InsecureSkipVerify is true, then the TLS client will be configured to ignore errors with certificates.
	//
	// This option is *not* recommended for production configurations.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// TLSSecurityProfile is the security profile to apply to the output connection
	TLSSecurityProfile *openshiftv1.TLSSecurityProfile `json:"securityProfile,omitempty"`
}

// OutputSecretSpec is a secret reference containing name only, no namespace.
type OutputSecretSpec struct {
	// Name of a secret in the namespace configured for log forwarder secrets.
	//
	// +required
	Name string `json:"name"`
}

// Filter defines a filter for log messages.
// See [FilterTypeSpec] for a list of filter types.
type FilterSpec struct {
	// Name used to refer to the filter from a `pipeline`.
	//
	// +kubebuilder:validation:minLength:=1
	// +required
	Name string `json:"name"`

	// Type of filter.
	//
	// +kubebuilder:validation:Enum:=kubeAPIAudit
	// +required
	Type string `json:"type"`

	FilterTypeSpec `json:",inline"`
}

// PipelinesSpec link a set of inputs to a set of outputs.
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

	// Filters lists the names of filters to be applied to records going through this pipeline.
	//
	// Each filter is applied in order.
	// If a filter drops a records, subsequent filters are not applied.
	// +optional
	FilterRefs []string `json:"filterRefs,omitempty"`

	// Labels applied to log records passing through this pipeline.
	// These labels appear in the `openshift.labels` map in the log record.
	//
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Name of the pipeline.
	//
	//+required
	Name string `json:"name"`

	// Parse enables parsing of log entries into structured logs
	//
	// Logs are parsed according to parse value, only `json` is supported as of now.
	//
	// +kubebuilder:validation:Enum:=json
	// +optional
	Parse string `json:"parse,omitempty"`

	// DetectMultilineErrors enables multiline error detection of container logs
	//
	// +optional
	DetectMultilineErrors bool `json:"detectMultilineErrors,omitempty"`
}

// CollectorSpec  defines scheduling and resources for the log collector.
type CollectorSpec struct {
	// The resource requirements for the collector
	//
	//+nullable
	//+optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	//
	//+nullable
	//+optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Node Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:ConfigMap"}
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Define the tolerations the Pods will accept
	//
	//+nullable
	//+optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Pod Tolerations",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Toleration"}
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

// ClusterLogForwarder a cluster-scoped resource that forwards logs for the entire cluster.
//
// Forwards application, infrastructure and audit logs.
//
// Configure forwarding by specifying a list of pipelines,
// which forward from a set of named inputs to a set of named outputs.
//
// For more details see the documentation on the API fields.
//
// +kubebuilder:object:root=true
// // +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=clf
// // +kubebuilder:storageversion
// +kubebuilder:skipversion
type ClusterLogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLogForwarderSpec   `json:"spec,omitempty"`
	Status ClusterLogForwarderStatus `json:"status,omitempty"`
}

// ClusterLogForwarderList contains a list of ClusterLogForwarder
//
// +kubebuilder:object:root=true
type ClusterLogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogForwarder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogForwarder{}, &ClusterLogForwarderList{})
}
