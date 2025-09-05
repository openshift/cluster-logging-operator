package context

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	kubernetes "sigs.k8s.io/controller-runtime/pkg/client"
)

// ForwarderContext provides the parameters needed to reconcile a log forwarder
type ForwarderContext struct {

	// Client is a read/write client for fetching resource
	Client kubernetes.Client

	// Reader is a read only client for retrieving kubernetes resources. This
	// client hits the API server directly, by-passing the controller cache
	Reader kubernetes.Reader

	// Forwarder is the ClusterLogForwarder to be reconciled
	Forwarder *obs.ClusterLogForwarder

	// Secrets a map of secrets spec'd by the forwarder
	Secrets map[string]*corev1.Secret

	// ConfigMaps specked by the forwarder
	ConfigMaps map[string]*corev1.ConfigMap

	// ClusterID is the unique ID of the cluster on which the operator is deployed
	ClusterID string

	// ClusterVersion is the version of the cluster on which the operator is deployed
	ClusterVersion string

	// AdditionalContext are additional context options to take pass along during reconciliation
	AdditionalContext utils.Options

	// Capabilities is the list of enabled features for the operator
	Capabilities Capabilities
}
