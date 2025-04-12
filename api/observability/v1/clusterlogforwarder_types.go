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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLogForwarderSpec defines the desired state of ClusterLogForwarder
type ClusterLogForwarderSpec struct {
	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator.
	//
	// +kubebuilder:default:=Managed
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Management State"
	ManagementState ManagementState `json:"managementState,omitempty"`

	// Specification of the Collector deployment to define
	// resource limits and workload placement
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Resources and Placement",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:advanced"}
	Collector *CollectorSpec `json:"collector,omitempty"`

	// Inputs are named filters for log messages to be forwarded.
	//
	// There are three built-in inputs named `application`, `infrastructure` and
	// `audit`. You don't need to define inputs here if those are sufficient for
	// your needs. See `inputRefs` for more.
	//
	// +kubebuilder:validation:Optional
	// +listType:=map
	// +listMapKey:=name
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Forwarder Inputs"
	Inputs []InputSpec `json:"inputs,omitempty"`

	// Outputs are named destinations for log messages.
	//
	// +kubebuilder:validation:Required
	// +listType:=map
	// +listMapKey:=name
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Forwarder Outputs"
	Outputs []OutputSpec `json:"outputs"`

	// Filters are applied to log records passing through a pipeline.
	// There are different types of filter that can select and modify log records in different ways.
	// See [FilterTypeSpec] for a list of filter types.
	//
	// +kubebuilder:validation:Optional
	// +listType:=map
	// +listMapKey:=name
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Forwarder Pipeline Filters"
	Filters []FilterSpec `json:"filters,omitempty"`

	// Pipelines forward the messages selected by a set of inputs to a set of outputs.
	//
	// +kubebuilder:validation:Required
	// +listType:=map
	// +listMapKey:=name
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Forwarder Pipelines"
	Pipelines []PipelineSpec `json:"pipelines"`

	// ServiceAccount points to the ServiceAccount resource used by the collector pods.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account"
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

type ServiceAccount struct {
	// Name of the ServiceAccount to use to deploy the Forwarder.  The ServiceAccount is created by the administrator
	//
	// +kubebuilder:validation:Pattern:="^[a-z][a-z0-9-]{2,62}[a-z0-9]$"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ServiceAccount Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Name string `json:"name"`
}

// ManagementState controls whether the operator's reconciliation is active for the given resource.
//
// +kubebuilder:validation:Enum:=Managed;Unmanaged
type ManagementState string

const (
	// ManagementStateManaged means that the operator is actively managing its operands and resources and driving them to meet the desired spec.
	ManagementStateManaged ManagementState = "Managed"

	// ManagementStateUnmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

// CollectorSpec is spec to define scheduling and resources for a collector
type CollectorSpec struct {
	// The resource requirements for the collector
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Collector Resource Requirements",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Define nodes for scheduling the pods.
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Node Selector"
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Define the tolerations the collector pods will accept
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tolerations"
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Define scheduling rules that influence pod placement based on node or pod affinity/anti-affinity constraints.
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Affinity"
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
}

// PipelineSpec links a set of inputs and transformations to a set of outputs.
type PipelineSpec struct {
	// Name of the pipeline
	//
	// +kubebuilder:validation:Pattern:="^[a-z][a-z0-9-]*[a-z0-9]$"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Name string `json:"name"`

	// InputRefs lists the names (`input.name`) of inputs to this pipeline.
	//
	// The following built-in input names are always available:
	//
	//  - `application` selects all logs from application pods.
	//
	//  - `infrastructure` selects logs from openshift and kubernetes pods and some node logs.
	//
	//  - `audit` selects node logs related to security audits.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems:=1
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Inputs"
	InputRefs []string `json:"inputRefs"`

	// OutputRefs lists the names (`output.name`) of outputs from this pipeline.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems:=1
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Outputs"
	OutputRefs []string `json:"outputRefs"`

	// Filters lists the names of filters to be applied to records going through this pipeline.
	//
	// Each filter is applied in order.
	// If a filter drops a records, subsequent filters are not applied.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Filters"
	FilterRefs []string `json:"filterRefs,omitempty"`
}

type LimitSpec struct {
	// MaxRecordsPerSecond is the maximum number of log records
	// allowed per input/output in a pipeline
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:validation:ExclusiveMinimum:=true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Max Records Per Second",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	MaxRecordsPerSecond int64 `json:"maxRecordsPerSecond"`
}

// ValueReference encodes a reference to a single field in either a ConfigMap or Secret in the same namespace.
//
// +kubebuilder:validation:XValidation:rule="has(self.configMapName) || has(self.secretName)", message="Either configMapName or secretName needs to be set"
// +kubebuilder:validation:XValidation:rule="!(has(self.configMapName) && has(self.secretName))", message="Only one of configMapName and secretName can be set"
type ValueReference struct {
	// Name of the key used to get the value in either the referenced ConfigMap or Secret.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Key Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Key string `json:"key"`

	// ConfigMapName contains the name of the ConfigMap containing the referenced value.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ConfigMap Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	ConfigMapName string `json:"configMapName,omitempty"`

	// SecretName contains the name of the Secret containing the referenced value.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	SecretName string `json:"secretName,omitempty"`
}

// SecretReference encodes a reference to a single key in a Secret in the same namespace.
type SecretReference struct {
	// Key contains the name of the key inside the referenced Secret.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Key Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Key string `json:"key"`

	// SecretName contains the name of the Secret containing the referenced value.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	SecretName string `json:"secretName"`
}

// BearerToken allows configuring the source of a bearer token used for authentication.
// The token can either be read from a secret or from a Kubernetes ServiceAccount.
// +kubebuilder:validation:XValidation:rule="self.from != 'secret' || has(self.secret)", message="Additional secret spec is required when bearer token is sourced from a secret"
type BearerToken struct {

	// From is the source from where to find the token
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Token Source"
	From BearerTokenFrom `json:"from"`

	// Use Secret if the value should be sourced from a Secret in the same namespace.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Token Secret"
	Secret *BearerTokenSecretKey `json:"secret,omitempty"`
}

// BearerTokenFrom specifies the source used for the bearer token.
//
// +kubebuilder:validation:Enum:=secret;serviceAccount
type BearerTokenFrom string

const (
	// BearerTokenFromSecret specifies to use the token from the spec'd secret
	BearerTokenFromSecret BearerTokenFrom = "secret"

	// BearerTokenFromServiceAccount specifies to use the token associated with the forwarder service account
	BearerTokenFromServiceAccount BearerTokenFrom = "serviceAccount"
)

type BearerTokenSecretKey struct {
	// Name of the key used to get the value from the referenced Secret.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Key Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Key string `json:"key"`

	// Name of secret
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Name string `json:"name"`
}

// TLSSpec contains options for TLS connections.
type TLSSpec struct {
	// CA can be used to specify a custom list of trusted certificate authorities.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Certificate Authority Bundle"
	CA *ValueReference `json:"ca,omitempty"`

	// Certificate points to the server certificate to use.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Certificate"
	Certificate *ValueReference `json:"certificate,omitempty"`

	// Key points to the private key of the server certificate.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Certificate Key"
	Key *SecretReference `json:"key,omitempty"`

	// KeyPassphrase points to the passphrase used to unlock the private key.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Certificate Key Passphrase"
	KeyPassphrase *SecretReference `json:"keyPassphrase,omitempty"`
}

// ClusterLogForwarderStatus defines the observed state of ClusterLogForwarder
type ClusterLogForwarderStatus struct {
	// Conditions of the log forwarder.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Forwarder Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// InputConditions maps input name to condition of the input.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Input Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	InputConditions []metav1.Condition `json:"inputConditions,omitempty"`

	// OutputConditions maps output name to condition of the output.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Output Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	OutputConditions []metav1.Condition `json:"outputConditions,omitempty"`

	// FilterConditions maps filter name to condition of the filter.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Filter Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	FilterConditions []metav1.Condition `json:"filterConditions,omitempty"`

	// PipelineConditions maps pipeline name to condition of the pipeline.
	//
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pipeline Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	PipelineConditions []metav1.Condition `json:"pipelineConditions,omitempty"`
}

// ClusterLogForwarder is an API to configure forwarding logs.
//
// You configure forwarding by specifying a list of `pipelines`,
// which forward from a set of named inputs to a set of named outputs.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=observability,shortName=obsclf;clf
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z][a-z0-9-]{1,61}[a-z0-9]$')",message="Name must be a valid DNS1035 label"
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

// FieldPath represents a path to find a value for a given field.  The format must be a value that can be converted to a
// valid collector configuration. It is a dot delimited path to a field in the log record. It must start with a `.`.
// The path can contain alphanumeric characters and underscores (a-zA-Z0-9_).
// If segments contain characters outside of this range, the segment must be quoted.
// Examples: `.kubernetes.namespace_name`, `.log_type`, '.kubernetes.labels.foobar', `.kubernetes.labels."foo-bar/baz"`
//
// +kubebuilder:validation:Pattern:=`^(\.[a-zA-Z0-9_]+|\."[^"]+")(\.[a-zA-Z0-9_]+|\."[^"]+")*$`
type FieldPath string

func init() {
	SchemeBuilder.Register(&ClusterLogForwarder{}, &ClusterLogForwarderList{})
}
