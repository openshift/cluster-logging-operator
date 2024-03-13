package v1

import (
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelSelector is a label query over a set of resources.
type LabelSelector metav1.LabelSelector

// Application log selector.
// All conditions in the selector must be satisfied (logical AND) to select logs.
type Application struct {
	// Namespaces from which to collect application logs.
	// Only messages from these namespaces are collected.
	// If absent or empty, logs are collected from all namespaces. This field supports
	// globs (e.g. mynam*space, *myanmespace)
	// Deprecated: Use []NamespaceContainerSpec instead.
	//
	// +optional
	// +deprecated
	Namespaces []string `json:"namespaces,omitempty"`

	// Selector for logs from pods with matching labels.
	// Only messages from pods with these labels are collected.
	// If absent or empty, logs are collected regardless of labels.
	//
	// +optional
	Selector *LabelSelector `json:"selector,omitempty"`

	// Group limit applied to the aggregated log
	// flow to this input. The total log flow from this input
	// cannot exceed the limit. Unsupported
	//
	// +optional
	// +docgen:ignore
	//+operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:hidden"}
	GroupLimit *LimitSpec `json:"-"` //`json:"groupLimit,omitempty"`

	// Container limit applied to each container of the pod(s) selected
	// by this input. No container of pods on selected by this input can
	// exceed this limit.  This limit is applied per collector deployment.
	//
	// +optional
	ContainerLimit *LimitSpec `json:"containerLimit,omitempty"`

	// Includes is the set of namespaces and containers to include when collecting logs.
	// Note: infrastructure namespaces are still excluded for "*" values unless a qualifying glob pattern is specified.
	//
	// +optional
	Includes []NamespaceContainerSpec `json:"includes,omitempty"`

	// Excludes is the set of namespaces and containers to ignore when collecting logs.
	// Takes precedence over Includes option.
	//
	// +optional
	Excludes []NamespaceContainerSpec `json:"excludes,omitempty"`
}

type NamespaceContainerSpec struct {

	// Namespace resources. Creates a combined file pattern together with Container resources.
	// Supports glob patterns and presumes "*" if ommitted.
	// Note: infrastructure namespaces are still excluded for "*" values unless a qualifying glob pattern is specified.
	//
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Container resources. Creates a combined file pattern together with Namespace resources.
	// Supports glob patterns and presumes "*" if ommitted.
	//
	// +optional
	Container string `json:"container,omitempty"`
}

// Infrastructure enables infrastructure logs. Filtering may be added in future.
// Sources of these logs:
// * container workloads deployed to namespaces: default, kube*, openshift*
// * journald logs from cluster nodes
type Infrastructure struct {

	// Sources defines the list of infrastructure sources to collect.
	// This field is optional and omission results in the collection of all infrastructure sources. Valid sources are:
	// node, container
	//
	// +optional
	Sources []string `json:"sources,omitempty"`
}

const (

	// InfrastructureSourceNode are journald logs from the node
	InfrastructureSourceNode string = "node"

	// InfrastructureSourceContainer are container logs from workloads deployed
	// in any of the following namespaces: default, kube*, openshift*
	InfrastructureSourceContainer string = "container"
)

var InfrastructureSources = sets.NewString(InfrastructureSourceNode, InfrastructureSourceContainer)

// Audit enables audit logs. Filtering may be added in future.
type Audit struct {
	// Sources defines the list of audit sources to collect.
	// This field is optional and its exclusion results in the collection of all audit sources. Valid sources are:
	// kubeAPI, openshiftAPI, auditd, ovn
	//
	// +optional
	Sources []string `json:"sources,omitempty"`
}

const (

	// AuditSourceKube are audit logs from kubernetes API servers
	AuditSourceKube string = "kubeAPI"

	// AuditSourceOpenShift are audit logs from OpenShift API servers
	AuditSourceOpenShift string = "openshiftAPI"

	// AuditSourceAuditd are audit logs from a node auditd service
	AuditSourceAuditd string = "auditd"

	// AuditSourceOVN are audit logs from an Open Virtual Network service
	AuditSourceOVN string = "ovn"
)

var AuditSources = sets.NewString(AuditSourceKube, AuditSourceOpenShift, AuditSourceAuditd, AuditSourceOVN)
