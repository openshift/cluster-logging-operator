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

package v1

import (
	openshiftv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ClusterLogForwarderKind = "ClusterLogForwarder"

// ClusterLogForwarderSpec defines how logs should be forwarded to remote targets.
type ClusterLogForwarderSpec struct {
	// Inputs are named filters for log messages to be forwarded.
	//
	// There are three built-in inputs named `application`, `infrastructure` and
	// `audit`. You don't need to define inputs here if those are sufficient for
	// your needs. See `inputRefs` for more.
	//
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Inputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderInputs"}
	Inputs []InputSpec `json:"inputs,omitempty"`

	// Outputs are named destinations for log messages.
	//
	// There is a built-in output named `default` which forwards to the default
	// openshift log store. You can define outputs to forward to other stores or
	// log processors, inside or outside the cluster.
	//
	// +optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Outputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderOutputs"}
	Outputs []OutputSpec `json:"outputs,omitempty"`

	// Pipelines forward the messages selected by a set of inputs to a set of outputs.
	//
	// +required
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Pipelines",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderPipelines"}
	Pipelines []PipelineSpec `json:"pipelines,omitempty"`

	// DEPRECATED OutputDefaults specify forwarder config explicitly for the
	// default managed log store named 'default'.  If there is a need to spec
	// the managed logstore, define an outputSpec like the following where the
	// managed fields (e.g. URL, Secret.Name) will be replaced with the required values:
	// spec:
	//   - outputs:
	//     - name: default
	//       type: elasticsearch
	//       elasticsearch:
	//         structuredTypeKey: kubernetes.labels.myvalue
	//
	// +optional
	OutputDefaults *OutputDefaults `json:"outputDefaults,omitempty"`
}

// ClusterLogForwarderStatus defines the observed state of ClusterLogForwarder
type ClusterLogForwarderStatus struct {
	// Conditions of the log forwarder.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Forwarder Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderConditions"}
	Conditions status.Conditions `json:"conditions,omitempty"`

	// Inputs maps input name to condition of the input.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Input Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:inputConditions"}
	Inputs NamedConditions `json:"inputs,omitempty"`

	// Outputs maps output name to condition of the output.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Output Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:outputConditions"}
	Outputs NamedConditions `json:"outputs,omitempty"`

	// Pipelines maps pipeline name to condition of the pipeline.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pipeline Conditions",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:pipelineConditions"}
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

	// Application, if present, enables named set of `application` logs that
	// can specify a set of match criteria
	//
	// +optional
	Application *Application `json:"application,omitempty"`

	// Infrastructure, if present, enables `infrastructure` logs.
	//
	// +optional
	// +docgen:ignore
	Infrastructure *Infrastructure `json:"infrastructure,omitempty"`

	// Audit, if present, enables `audit` logs.
	//
	// +optional
	// +docgen:ignore
	Audit *Audit `json:"audit,omitempty"`
}

// NOTE: We currently only support matchLabels so define a LabelSelector type with
// only matchLabels. When matchExpressions is implemented (LOG-1126), replace this with:
// k8s.io/apimachinery/pkg/apis/meta/v1.LabelSelector

// A label selector is a label query over a set of resources.
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

// Infrastructure enables infrastructure logs. Filtering may be added in future.
// +docgen:ignore
type Infrastructure struct{}

// Audit enables audit logs. Filtering may be added in future.
// +docgen:ignore
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

type OutputDefaults struct {

	// Elasticsearch OutputSpec default values
	//
	// Values specified here will be used as default values for Elasticsearch Output spec
	//
	// +kubebuilder:default:false
	// +optional
	Elasticsearch *ElasticsearchStructuredSpec `json:"elasticsearch,omitempty"`
}

// +kubebuilder:object:root=true
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
type ClusterLogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of ClusterLogForwarder
	Spec ClusterLogForwarderSpec `json:"spec,omitempty"`

	// Status of the ClusterLogForwarder
	Status ClusterLogForwarderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ClusterLogForwarderList contains a list of ClusterLogForwarder
type ClusterLogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogForwarder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogForwarder{}, &ClusterLogForwarderList{})
}
