package k8shandler

import (
	"context"
	"errors"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/types"
)

func Reconcile(requestCluster *logging.ClusterLogging, requestClient client.Client) (err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Client:  requestClient,
		Cluster: requestCluster,
	}

	forwarder := clusterLoggingRequest.getLogForwarder()
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	proxyConfig := clusterLoggingRequest.getProxyConfig()

	// Reconcile certs
	if err = clusterLoggingRequest.CreateOrUpdateCertificates(); err != nil {
		return fmt.Errorf("Unable to create or update certificates for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile Log Store
	if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
		return fmt.Errorf("Unable to create or update logstore for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile Visualization
	if err = clusterLoggingRequest.CreateOrUpdateVisualization(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update visualization for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile Curation
	if err = clusterLoggingRequest.CreateOrUpdateCuration(); err != nil {
		return fmt.Errorf("Unable to create or update curation for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile Metrics Dashboards
	if err = clusterLoggingRequest.CreateOrUpdateDashboards(); err != nil {
		return fmt.Errorf("Unable to create or update metrics dashboards for %q: %w", clusterLoggingRequest.Cluster.Name, err)
	}

	return nil
}

func ReconcileForClusterLogForwarder(forwarder *logging.ClusterLogForwarder, requestClient client.Client) (err error) {
	logger.DebugObject("Reconciling ClusterLogForwarder instance: %v", forwarder)
	clusterLoggingRequest := ClusterLoggingRequest{
		Client: requestClient,
	}
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	clusterLogging := clusterLoggingRequest.getClusterLogging()
	if clusterLogging == nil {
		return nil
	}
	clusterLoggingRequest.Cluster = clusterLogging

	if clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		return nil
	}

	proxyConfig := clusterLoggingRequest.getProxyConfig()

	// Reconcile Collection
	err = clusterLoggingRequest.CreateOrUpdateCollection(proxyConfig)
	forwarder.Status = clusterLoggingRequest.ForwarderRequest.Status
	logger.DebugObject("ClusterLogForwarder status after updating collection: %#v", forwarder.Status)
	if err != nil {
		msg := fmt.Sprintf("Unable to reconcile collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		logger.Errorf(msg)
		return errors.New(msg)
	}
	return nil
}

func ReconcileForGlobalProxy(proxyConfig *configv1.Proxy, requestClient client.Client) (err error) {

	clusterLoggingRequest := ClusterLoggingRequest{
		Client: requestClient,
	}

	clusterLogging := clusterLoggingRequest.getClusterLogging()
	if clusterLogging == nil {
		return nil
	}

	clusterLoggingRequest.Cluster = clusterLogging

	if clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		return nil
	}

	forwarder := clusterLoggingRequest.getLogForwarder()
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(proxyConfig); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	return nil
}

func ReconcileForTrustedCABundle(requestName string, requestClient client.Client) (err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Client: requestClient,
	}

	clusterLogging := clusterLoggingRequest.getClusterLogging()
	if clusterLogging == nil {
		return nil
	}

	clusterLoggingRequest.Cluster = clusterLogging

	if clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		return nil
	}

	forwarder := clusterLoggingRequest.getLogForwarder()
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	proxyConfig := clusterLoggingRequest.getProxyConfig()

	return clusterLoggingRequest.RestartFluentd(proxyConfig)
}

func (clusterRequest *ClusterLoggingRequest) getClusterLogging() *logging.ClusterLogging {
	clusterLoggingNamespacedName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.OpenshiftNS}
	clusterLogging := &logging.ClusterLogging{}

	if err := clusterRequest.Client.Get(context.TODO(), clusterLoggingNamespacedName, clusterLogging); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Printf("Encountered unexpected error getting %v", clusterLoggingNamespacedName)
		}
		return nil
	}

	return clusterLogging
}

func (clusterRequest *ClusterLoggingRequest) getProxyConfig() *configv1.Proxy {
	// we need to see if we have the proxy available so we
	// don't blank out any proxy configured changes...
	proxyNamespacedName := types.NamespacedName{Name: constants.ProxyName}
	proxyConfig := &configv1.Proxy{}
	if err := clusterRequest.Client.Get(context.TODO(), proxyNamespacedName, proxyConfig); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Printf("Encountered unexpected error getting %v", proxyNamespacedName)
		}
	}

	return proxyConfig
}

func (clusterRequest *ClusterLoggingRequest) getLogForwarder() *logging.ClusterLogForwarder {
	name := types.NamespacedName{Name: constants.SingletonName} // Cluster scope, no namespace
	forwarder := &logging.ClusterLogForwarder{}
	logger.Debug("clusterlogforwarder-controller fetching LF instance")
	if err := clusterRequest.Client.Get(context.TODO(), name, forwarder); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Printf("Encountered unexpected error getting %v", name)
		}
	}

	return forwarder
}
