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

package v2alpha1

import (
	"github.com/openshift/cluster-logging-operator/internal/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const LogForwarderKind = "LogForwarder"

// LogForwarder forwards container logs for a single namespace.
//
// You configure forwarding by specifying a list of pipelines,
// which forward from a set of named inputs to a set of named outputs.
//
// For more details see the documentation on the API fields.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=lf
type LogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogForwarderSpec   `json:"spec,omitempty"`
	Status LogForwarderStatus `json:"status,omitempty"`
}

// LogForwarderList contains a list of LogForwarder
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type LogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogForwarder `json:"items"`
}

// LogForwarderSpec specifies log forwarding for a single namespace.
type LogForwarderSpec struct {
	// Inputs define selection criteria for logs to be forwarded.
	//
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Inputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderInputs"}
	Inputs []InputSpec `json:"inputs,omitempty"`

	LogForwarderCommonSpec `json:",inline"`
}

// LogForwarderCommonSpec is shared between LogForwarderSpec and ClusterLogForwarderSpec
type LogForwarderCommonSpec struct {

	// Outputs define destinations for log messages.
	//
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Outputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderOutputs"}
	Outputs []OutputSpec `json:"outputs,omitempty"`

	// Pipelines forward the messages from a set of named inputs to a set of named outputs.
	//
	// +required
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Pipelines",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderPipelines"}
	Pipelines []PipelineSpec `json:"pipelines,omitempty"`

	// FIXME Suggest dropping CollectorRef and automatically associating with a LogCollector in the same namespace, with the same name.

	// CollectorRef is the name of a LogCollector resource in the same namespace.
	// It provides additional configuration for log collection by this forwarder.
	//
	//+optional
	CollectorRef string `json:"collectorRef,omitempty"`
}

// LogForwarderStatus defines the observed state of LogForwarder
type LogForwarderStatus struct {
	// Conditions of the log forwarder.
	//
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Forwarder Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderConditions"}
	Conditions status.Conditions `json:"conditions,omitempty"`

	// Inputs maps input name to the condition of the input.
	//
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Input Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:inputConditions"}
	Inputs NamedConditions `json:"inputs,omitempty"`

	// Outputs maps output name to the condition of the output.
	//
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Output Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:outputConditions"}
	Outputs NamedConditions `json:"outputs,omitempty"`

	// Pipelines maps pipeline name to the condition of the pipeline.
	//
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pipeline Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:pipelineConditions"}
	Pipelines NamedConditions `json:"pipelines,omitempty"`
}

// inputs defines selection criteria for logs to be forwarded.
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +kubebuilder:validation:minLength:=1
	// +required
	Name string `json:"name"`

	// NOTE: the following fields in this struct are deliberately _not_ `omitempty`.
	// An empty field means enable that input type with no filter.

	// Application enables container logs that can meet the selectino criteria
	//
	// +optional
	Application *Application `json:"application,omitempty"`
}

// NOTE: We currently only support matchLabels so define a LabelSelector type with
// only matchLabels. When matchExpressions is implemented (LOG-1126), replace this with:
// k8s.io/apimachinery/pkg/apis/meta/LabelSelector

// LabelSelector selects logs from pods with matching labels.
type LabelSelector struct {
	// matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
	// map is equivalent to an element of matchExpressions, whose key field is "key", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty" protobuf:"bytes,1,rep,name=matchLabels"`
}

// Application log selector.
// All conditions in the selector must be satisfied (logical AND) to select logs.
type Application struct {
	// Namespaces from which to collect application logs.
	// Only messages from these namespaces are collected.
	// If absent or empty, logs are collected from all namespaces.
	//
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// Selector for logs from pods with matching labels.
	// Only messages from pods with these labels are collected.
	// If absent or empty, logs are collected regardless of labels.
	//
	// +optional
	Selector *LabelSelector `json:"selector,omitempty"`
}

// OutputSpec defines a destination for log messages.
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
	// Names a secret in the same namespace as the LogForwarder.
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
}

// OutputTLSSpec contains options for TLS connections that are agnostic to the output type.
type OutputTLSSpec struct {
	// If InsecureSkipVerify is true, then the TLS client will be configured to ignore errors with certificates.
	//
	// This option is *not* recommended for production configurations.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// OutputSecretSpec is a secret reference containing name only, no namespace.
type OutputSecretSpec struct {
	// Name of a secret in the namespace configured for log forwarder secrets.
	//
	// +required
	Name string `json:"name"`
}

type PipelineSpec struct {
	// InputRefs lists the names (`input.name`) of inputs to this pipeline.
	//
	// The built-in input name `application` is always available.
	//
	// +required
	InputRefs []string `json:"inputRefs"`

	PipelineCommonSpec `json:",inline"`
}

// PipelineCommonSpec is shared by LogForwarderSpec and ClusterLogForwarderSpec
type PipelineCommonSpec struct {
	// OutputRefs lists the names (`output.name`) of outputs from this pipeline.
	//
	// The following built-in names are always available:
	//
	// 'default' Output to the default log store provided by ClusterLogging.
	//
	// +required
	OutputRefs []string `json:"outputRefs"`

	// Labels applied to log records passing through this pipeline.
	// These labels appear in the `openshift.labels` map in the log record.
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

	// DetectMultilineErrors enables multiline error detection of container logs
	//
	// +optional
	DetectMultilineErrors bool `json:"detectMultilineErrors,omitempty"`
}

func init() {
	SchemeBuilder.Register(&LogForwarder{}, &LogForwarderList{})
}
