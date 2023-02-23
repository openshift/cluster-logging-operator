package v1

import (
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	eventv1 "k8s.io/api/core/v1"
)

// The data model for collected logs from containers.
//
// +kubebuilder:object:root=true
// +docgen:displayname=Viaq Data Model for Containers
type ContainerLog types.ContainerLog

// The data model for collected logs from node journal.
//
// +kubebuilder:object:root=true
// +docgen:displayname=Viaq Data Model for journald
type JournalLog types.JournalLog

// The data model for collected audit event logs from kubernetes or OpenShift api servers.
//
// +kubebuilder:object:root=true
// +docgen:displayname=Viaq Data Model for kubernetes api events
// nolint:govet
type ApiEvent struct {
	types.ViaQCommon `json:",inline,omitempty"`
	eventv1.Event    `json:",inline,omitempty"`
}
