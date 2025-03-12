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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InputType specifies the type of log input to create.
//
// +kubebuilder:validation:Enum:=audit;application;infrastructure;receiver
type InputType string

const (
	// InputTypeApplication contains all the non-infrastructure container logs.
	InputTypeApplication InputType = "application"
	// InputTypeInfrastructure contains infrastructure containers and system logs.
	InputTypeInfrastructure InputType = "infrastructure"
	// InputTypeAudit contains system audit logs.
	InputTypeAudit InputType = "audit"
	// InputTypeReceiver defines a network receiver for receiving logs from non-cluster sources.
	InputTypeReceiver InputType = "receiver"
)

var (
	InputTypes = []InputType{
		InputTypeApplication,
		InputTypeInfrastructure,
		InputTypeAudit,
		InputTypeReceiver,
	}
)

// InputSpec defines a selector of log messages for a given log type.
// +kubebuilder:validation:XValidation:rule="self.type != 'application' || has(self.application)", message="Additional type specific spec is required for the input type"
// +kubebuilder:validation:XValidation:rule="self.type != 'infrastructure' || has(self.infrastructure)", message="Additional type specific spec is required for the input type"
// +kubebuilder:validation:XValidation:rule="self.type != 'audit' || has(self.audit)", message="Additional type specific spec is required for the input type"
// +kubebuilder:validation:XValidation:rule="self.type != 'receiver' || has(self.receiver)", message="Additional type specific spec is required for the input type"
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +kubebuilder:validation:Pattern:="^[a-z][a-z0-9-]*[a-z0-9]$"
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Input Name"
	Name string `json:"name"`

	// Type of output sink.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Input Type"
	Type InputType `json:"type"`

	// Application, named set of `application` logs that
	// can specify a set of match criteria
	//
	// +nullable
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Application Logs Input"
	Application *Application `json:"application,omitempty"`

	// Infrastructure, Enables `infrastructure` logs.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Infrastructure Logs Input"
	Infrastructure *Infrastructure `json:"infrastructure,omitempty"`

	// Audit, enables `audit` logs.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Audit Logs Input"
	Audit *Audit `json:"audit,omitempty"`

	// Receiver to receive logs from non-cluster sources.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Receiver"
	Receiver *ReceiverSpec `json:"receiver,omitempty"`
}

type ContainerInputTuningSpec struct {

	// RateLimitPerContainer is the limit applied to each container
	// by this input. This limit is applied per collector deployment.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Per-Container Rate Limit"
	RateLimitPerContainer *LimitSpec `json:"rateLimitPerContainer,omitempty"`
}

// ApplicationSource defines the type of ApplicationSource log source to use.
//
// +kubebuilder:validation:Enum:=container
type ApplicationSource string

const (

	// ApplicationSourceContainer are container logs from deployed workloads
	// in any of the following namespaces: default, kube*, openshift*
	ApplicationSourceContainer ApplicationSource = "container"
)

// Application workload log selector.
// All conditions in the selector must be satisfied (logical AND) to select logs.
type Application struct {
	// Selector for logs from pods with matching labels.
	//
	// Only messages from pods with these labels are collected.
	//
	// If absent or empty, logs are collected regardless of labels.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pod Selector",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Pod"}
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Tuning is the container input tuning spec for this container sources
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Input Tuning"
	Tuning *ContainerInputTuningSpec `json:"tuning,omitempty"`

	// Includes is the set of namespaces and containers to include when collecting logs.
	//
	// Note: infrastructure namespaces are still excluded for "*" values unless a qualifying glob pattern is specified.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Include"
	Includes []NamespaceContainerSpec `json:"includes,omitempty"`

	// Excludes is the set of namespaces and containers to ignore when collecting logs.
	//
	// Takes precedence over Includes option.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Exclude"
	Excludes []NamespaceContainerSpec `json:"excludes,omitempty"`
}

type NamespaceContainerSpec struct {

	// Namespace specs the namespace from which to collect logs
	// Supports glob patterns and presumes "*" if omitted.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Namespace Glob",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Namespace string `json:"namespace,omitempty"`

	// Container spec the containers from which to collect logs
	// Supports glob patterns and presumes "*" if omitted.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Glob",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Container string `json:"container,omitempty"`
}

// InfrastructureSource defines the type of infrastructure log source to use.
//
// +kubebuilder:validation:Enum:=container;node
type InfrastructureSource string

const (
	// InfrastructureSourceNode are journald logs from the node
	InfrastructureSourceNode InfrastructureSource = "node"

	// InfrastructureSourceContainer are container logs from workloads deployed
	// in any of the following namespaces: default, kube*, openshift*
	InfrastructureSourceContainer InfrastructureSource = "container"
)

var (
	InfrastructureSources = []InfrastructureSource{
		InfrastructureSourceNode,
		InfrastructureSourceContainer,
	}
)

// Infrastructure enables infrastructure logs.
// Sources of these logs:
// * container workloads deployed to namespaces: default, kube*, openshift*
// * journald logs from cluster nodes
type Infrastructure struct {
	// Sources defines the list of infrastructure sources to collect.
	// This field is optional and omission results in the collection of all infrastructure sources.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Sources"
	Sources []InfrastructureSource `json:"sources,omitempty"`
}

// AuditSource defines which type of audit log source is used.
//
// +kubebuilder:validation:Enum:=auditd;kubeAPI;openshiftAPI;ovn
type AuditSource string

const (
	// AuditSourceKube are audit logs from kubernetes API servers
	AuditSourceKube AuditSource = "kubeAPI"

	// AuditSourceOpenShift are audit logs from OpenShift API servers
	AuditSourceOpenShift AuditSource = "openshiftAPI"

	// AuditSourceAuditd are audit logs from a node auditd service
	AuditSourceAuditd AuditSource = "auditd"

	// AuditSourceOVN are audit logs from an Open Virtual Network service
	AuditSourceOVN AuditSource = "ovn"
)

var (
	AuditSources = []AuditSource{
		AuditSourceKube,
		AuditSourceOpenShift,
		AuditSourceAuditd,
		AuditSourceOVN,
	}
)

// Audit enables audit logs.
type Audit struct {
	// Sources defines the list of audit sources to collect.
	// This field is optional and its exclusion results in the collection of all audit sources.
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Sources"
	Sources []AuditSource `json:"sources,omitempty"`
}

// ReceiverType specifies the type of receiver that should be created.
//
// +kubebuilder:validation:Enum:=http;syslog
type ReceiverType string

const (
	ReceiverTypeHTTP   ReceiverType = "http"
	ReceiverTypeSyslog ReceiverType = "syslog"
)

var (
	ReceiverTypes = []ReceiverType{
		ReceiverTypeHTTP,
		ReceiverTypeSyslog,
	}
)

type InputTLSSpec TLSSpec

// ReceiverSpec is a union of input Receiver types.
type ReceiverSpec struct {
	// Type of Receiver plugin.
	//
	// Supported Receiver types are:
	//
	// 1. http
	//    - Currently only supports kubernetes audit logs (log_type = "audit")
	// 2. syslog
	//    - Currently only supports node infrastucture logs (log_type = "infrastructure")
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Receiver Type"
	Type ReceiverType `json:"type"`

	// TLS contains settings for controlling options of TLS connections.
	//
	// The operator will request certificates from the cluster's cert signing service when TLS is not defined.
	// The certificates are injected into a secret named "<clusterlogforwarder.name>-<input.name>" which is mounted into
	// the collector. The collector is configured to use the public and private key provided by the service
	//
	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS Options"
	TLS *InputTLSSpec `json:"tls,omitempty"`

	// Port the Receiver listens on. It must be a value between 1024 and 65535
	//
	// +kubebuilder:validation:Minimum:=1024
	// +kubebuilder:validation:Maximum:=65535
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Listen Port",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	Port int32 `json:"port"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="HTTP Receiver Configuration"
	HTTP *HTTPReceiver `json:"http,omitempty"`
}

// HTTPReceiverFormat defines the type of log data incoming through the HTTP receiver.
//
// +kubebuilder:validation:Enum:=kubeAPIAudit
type HTTPReceiverFormat string

const (
	HTTPReceiverFormatKubeAPIAudit HTTPReceiverFormat = "kubeAPIAudit"
)

// HTTPReceiver receives encoded logs as a HTTP endpoint.
type HTTPReceiver struct {
	// Format is the format of incoming log data.
	//
	// The only currently supported format is `kubeAPIAudit`.
	//
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Data Format"
	Format HTTPReceiverFormat `json:"format"`
}
