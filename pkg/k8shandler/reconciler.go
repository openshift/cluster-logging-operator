package k8shandler

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logforwarding "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/types"
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

	// we need to see if we have the proxy available so we
	// don't blank out any proxy configured changes...
	proxyNamespacedName := types.NamespacedName{Name: constants.ProxyName}
	proxyConfig := &configv1.Proxy{}
	if err := clusterLoggingRequest.client.Get(context.TODO(), proxyNamespacedName, proxyConfig); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Errorf("Encountered unexpected error getting %v", proxyNamespacedName)
		}
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
	if err = clusterLoggingRequest.CreateOrUpdateVisualization(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update visualization for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Curation
	if err = clusterLoggingRequest.CreateOrUpdateCuration(); err != nil {
		return fmt.Errorf("Unable to create or update curation for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	return nil
}

func ReconcileForGlobalProxy(requestCluster *logging.ClusterLogging, forwarding *logforwarding.LogForwarding, proxyConfig *configv1.Proxy, requestClient client.Client) (err error) {

	clusterLoggingRequest := ClusterLoggingRequest{
		client:            requestClient,
		cluster:           requestCluster,
		ForwardingRequest: forwarding,
	}
	if forwarding != nil {
		clusterLoggingRequest.ForwardingSpec = forwarding.Spec
	}

	// Reconcile Visualization
	if err = clusterLoggingRequest.CreateOrUpdateVisualization(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update visualization for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.cluster.Name, err)
	}

	return nil
}

func ReconcileForKibanaSecret(requestClient client.Client) (err error) {

	clusterLoggingRequest := ClusterLoggingRequest{
		client: requestClient,
	}

	clusterLogging := clusterLoggingRequest.getClusterLogging()
	clusterLoggingRequest.cluster = clusterLogging

	if clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		return nil
	}

	// call for Kibana to restart itself (e.g. delete its pods)
	return clusterLoggingRequest.RestartKibana()
}

func (clusterRequest *ClusterLoggingRequest) getClusterLogging() *logging.ClusterLogging {
	clusterLoggingNamespacedName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.OpenshiftNS}
	clusterLogging := &logging.ClusterLogging{}

	if err := clusterRequest.client.Get(context.TODO(), clusterLoggingNamespacedName, clusterLogging); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Printf("Encountered unexpected error getting %v", clusterLoggingNamespacedName)
		}
	}

	return clusterLogging
}

func (clusterRequest *ClusterLoggingRequest) getProxyConfig() *configv1.Proxy {
	// we need to see if we have the proxy available so we
	// don't blank out any proxy configured changes...
	proxyNamespacedName := types.NamespacedName{Name: constants.ProxyName}
	proxyConfig := &configv1.Proxy{}
	if err := clusterRequest.client.Get(context.TODO(), proxyNamespacedName, proxyConfig); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Printf("Encountered unexpected error getting %v", proxyNamespacedName)
		}
	}

	return proxyConfig
}

func (clusterRequest *ClusterLoggingRequest) getLogForwarding() *logforwarding.LogForwarding {
	logForwardingNamespacedName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.OpenshiftNS}
	logForwarding := &logforwarding.LogForwarding{}
	logger.Debug("logforwarding-controller fetching LF instance")
	if err := clusterRequest.client.Get(context.TODO(), logForwardingNamespacedName, logForwarding); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Printf("Encountered unexpected error getting %v", logForwardingNamespacedName)
		}
	}

	return logForwarding
}
