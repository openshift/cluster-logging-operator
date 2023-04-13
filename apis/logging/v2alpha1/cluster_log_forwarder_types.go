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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ClusterLogForwarderKind = "ClusterLogForwarder"

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
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=clf
type ClusterLogForwarder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLogForwarderSpec `json:"spec,omitempty"`
	Status LogForwarderStatus      `json:"status,omitempty"`
}

// ClusterLogForwarderList contains a list of ClusterLogForwarder
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
type ClusterLogForwarderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogForwarder `json:"items"`
}

// ClusterLogForwarderSpec specifies log forwarding for the entire cluster.
type ClusterLogForwarderSpec struct {
	// Inputs are named filters for log messages to be forwarded.
	//
	// There are three built-in inputs named `application`, `infrastructure` and
	// `audit`. You don't need to define inputs here if those are sufficient for
	// your needs. See `inputRefs` for more.
	//
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Forwarder Inputs",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:forwarderInputs"}
	Inputs []ClusterInputSpec `json:"inputs,omitempty"`

	LogForwarderCommonSpec `json:",inline"`
}

// input defines selection criteria for logs to be forwarded.
type ClusterInputSpec struct {
	InputSpec `json:",inline"`

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

// Infrastructure enables infrastructure logs. Filtering may be added in future.
// +docgen:ignore
type Infrastructure struct{}

// Audit enables audit logs. Filtering may be added in future.
// +docgen:ignore
type Audit struct{}

type ClusterPipelineSpec struct {
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
	ClusterInputRefs []string `json:"inputRefs"`

	PipelineCommonSpec `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogForwarder{}, &ClusterLogForwarderList{})
}
