package k8shandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViaQ/logerr/log"
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/types"
)

func Reconcile(requestCluster *logging.ClusterLogging, requestClient client.Client, r record.EventRecorder) (err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Client:        requestClient,
		Cluster:       requestCluster,
		EventRecorder: r,
	}

	if !clusterLoggingRequest.isManaged() {
		return nil
	}

	forwarder := clusterLoggingRequest.getLogForwarder()
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	if clusterLoggingRequest.IncludesManagedStorage() {
		// Reconcile certs
		if err = clusterLoggingRequest.CreateOrUpdateCertificates(); err != nil {
			return fmt.Errorf("Unable to create or update certificates for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

		// Reconcile Log Store
		if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
			return fmt.Errorf("Unable to create or update logstore for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

		// Reconcile Visualization
		if err = clusterLoggingRequest.CreateOrUpdateVisualization(); err != nil {
			return fmt.Errorf("Unable to create or update visualization for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

	} else {
		removeManagedStorage(clusterLoggingRequest)
	}

	// Remove Curator
	if err := clusterLoggingRequest.removeCurator(); err != nil {
		log.V(1).Error(err, "Error removing curator component")
	}
	clusterLoggingRequest.Cluster.Status.Conditions.SetCondition(status.Condition{
		Type:    "CuratorRemoved",
		Status:  corev1.ConditionTrue,
		Reason:  "ResourceDeprecated",
		Message: "curator is deprecated in favor of defining retention policy",
	})

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(); err != nil {
		return fmt.Errorf("Unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile Metrics Dashboards
	if err = clusterLoggingRequest.CreateOrUpdateDashboards(); err != nil {
		return fmt.Errorf("Unable to create or update metrics dashboards for %q: %w", clusterLoggingRequest.Cluster.Name, err)
	}

	return nil
}

func removeManagedStorage(clusterRequest ClusterLoggingRequest) {
	log.V(1).Info("Removing managed store components...")
	for _, remove := range []func() error{clusterRequest.removeElasticsearch, clusterRequest.removeKibana} {
		if err := remove(); err != nil {
			log.V(1).Error(err, "Error removing component")
		}
	}
}

func ReconcileForClusterLogForwarder(forwarder *logging.ClusterLogForwarder, requestClient client.Client) (err error) {
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

	// Reconcile Collection
	err = clusterLoggingRequest.CreateOrUpdateCollection()
	forwarder.Status = clusterLoggingRequest.ForwarderRequest.Status
	if err != nil {
		msg := fmt.Sprintf("Unable to reconcile collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		log.Error(err, msg)
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
	if err = clusterLoggingRequest.CreateOrUpdateCollection(); err != nil {
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

	return clusterLoggingRequest.RestartFluentd()
}

func (clusterRequest *ClusterLoggingRequest) getClusterLogging() *logging.ClusterLogging {
	clusterLoggingNamespacedName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.OpenshiftNS}
	clusterLogging := &logging.ClusterLogging{}

	if err := clusterRequest.Client.Get(context.TODO(), clusterLoggingNamespacedName, clusterLogging); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "NamespacedName", clusterLoggingNamespacedName)
		}
		return nil
	}

	return clusterLogging
}

func (clusterRequest *ClusterLoggingRequest) getLogForwarder() *logging.ClusterLogForwarder {
	nsname := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.OpenshiftNS}
	forwarder := &logging.ClusterLogForwarder{}
	if err := clusterRequest.Client.Get(context.TODO(), nsname, forwarder); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "forwarder", nsname)
		}
	}

	return forwarder
}
