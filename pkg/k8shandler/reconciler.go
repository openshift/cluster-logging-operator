package k8shandler

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logforwarding "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func Reconcile(requestCluster *logging.ClusterLogging, forwarding *logforwarding.LogForwarding, requestClient client.Client) (err error) {
	logger.Debugf("Reconciling cl: %v, forwarding: %v", requestCluster, forwarding)
	clusterLoggingRequest := ClusterLoggingRequest{
		client:            requestClient,
		cluster:           requestCluster,
		ForwardingRequest: forwarding,
	}
	if forwarding != nil {
		clusterLoggingRequest.ForwardingSpec = forwarding.Spec
	}

	// Reconcile certs
	if err = clusterLoggingRequest.CreateOrUpdateCertificates(); err != nil {
		return fmt.Errorf("Unable to create or update certificates for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Log Store
	if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
		return fmt.Errorf("Unable to create or update logstore for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Visualization
	if err = clusterLoggingRequest.CreateOrUpdateVisualization(nil, nil); err != nil {
		return fmt.Errorf("Unable to create or update visualization for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Curation
	if err = clusterLoggingRequest.CreateOrUpdateCuration(); err != nil {
		return fmt.Errorf("Unable to create or update curation for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(nil, nil); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	return nil
}

func ReconcileForGlobalProxy(requestCluster *logging.ClusterLogging, forwarding *logforwarding.LogForwarding, proxyConfig *configv1.Proxy, trustedCABundleCM *corev1.ConfigMap, requestClient client.Client) (err error) {

	clusterLoggingRequest := ClusterLoggingRequest{
		client:            requestClient,
		cluster:           requestCluster,
		ForwardingRequest: forwarding,
	}
	if forwarding != nil {
		clusterLoggingRequest.ForwardingSpec = forwarding.Spec
	}

	// Reconcile Visualization
	if err = clusterLoggingRequest.CreateOrUpdateVisualization(proxyConfig, trustedCABundleCM); err != nil {
		return fmt.Errorf("Unable to create or update visualization for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(proxyConfig, trustedCABundleCM); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	return nil
}
