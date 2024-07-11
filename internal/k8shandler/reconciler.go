package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func Reconcile(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string, resourceNames *factory.ForwarderResourceNames) (err error) {
	log.V(3).Info("Reconciling", "ClusterLogging", cl, "ClusterLogForwarder", forwarder)
	clusterLoggingRequest := NewClusterLoggingRequest(cl, forwarder, requestClient, reader, r, clusterVersion, clusterID, resourceNames)

	if !clusterLoggingRequest.isManaged() {
		return nil
	}

	if !forwarder.Status.IsReady() {
		removeCollectorAndUpdate(clusterLoggingRequest)
		return nil
	}

	// Remove existing collector deployment/daemonset
	if clusterLoggingRequest.isDaemonset {
		if err := collector.RemoveDeployment(clusterLoggingRequest.Client, forwarder.Namespace, clusterLoggingRequest.ResourceNames.DaemonSetName()); err != nil {
			return err
		}
	} else {
		if err := collector.Remove(clusterLoggingRequest.Client, forwarder.Namespace, clusterLoggingRequest.ResourceNames.DaemonSetName()); err != nil {
			return err
		}
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(); err != nil {
		return fmt.Errorf("unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Clean up any stale http input services
	if err = clusterLoggingRequest.RemoveInputServices([]metav1.OwnerReference{utils.AsOwner(forwarder)}, false); err != nil {
		return fmt.Errorf("error removing stale http input services. %v", err)
	}

	return nil
}

func removeCollectorAndUpdate(clusterRequest ClusterLoggingRequest) {
	log.V(3).Info("forwarder not found and logStore not found so removing collector")
	if err := clusterRequest.removeCollector(); err != nil {
		log.Error(err, "Error removing collector")
	}
}
