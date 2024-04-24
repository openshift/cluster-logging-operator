package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	eslogstore "github.com/openshift/cluster-logging-operator/internal/logstore/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	logmetricexporter "github.com/openshift/cluster-logging-operator/internal/metrics/logfilemetricexporter"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func Reconcile(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string, resourceNames *factory.ForwarderResourceNames) (err error) {
	log.V(3).Info("Reconciling", "ClusterLogging", cl, "ClusterLogForwarder", forwarder)
	clusterLoggingRequest := NewClusterLoggingRequest(cl, forwarder, requestClient, reader, r, clusterVersion, clusterID, resourceNames)

	if !clusterLoggingRequest.isManaged() {
		return nil
	}

	if clusterLoggingRequest.IsLegacyDeployment() {
		if clusterLoggingRequest.IncludesManagedStorage() {
			// Reconcile Log Store
			if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
				return fmt.Errorf("unable to create or update logstore for %q: %v", clusterLoggingRequest.Cluster.Name, err)
			}

			// Reconcile Visualization
			if err = clusterLoggingRequest.CreateOrUpdateVisualization(); err != nil {
				return fmt.Errorf("unable to create or update visualization for %q: %v", clusterLoggingRequest.Cluster.Name, err)
			}

		} else {
			removeManagedStorage(clusterLoggingRequest)
		}
	}

	if !forwarder.Status.IsReady() {
		removeCollectorAndUpdate(clusterLoggingRequest)
		return nil
	}

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(); err != nil {
		telemetry.IncreaseCollectorErrors()
		return fmt.Errorf("unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Clean up any stale http input services
	if err = clusterLoggingRequest.RemoveInputServices([]metav1.OwnerReference{utils.AsOwner(forwarder)}, false); err != nil {
		return fmt.Errorf("error removing stale http input services")
	}

	telemetry.UpdateDefaultForwarderInfo(clusterLoggingRequest.Forwarder)
	return nil
}

func removeCollectorAndUpdate(clusterRequest ClusterLoggingRequest) {
	log.V(3).Info("forwarder not found and logStore not found so removing collector")
	if err := clusterRequest.removeCollector(); err != nil {
		log.Error(err, "Error removing collector")
	}
}

func removeManagedStorage(clusterRequest ClusterLoggingRequest) {
	log.V(1).Info("Removing managed store components...")
	for _, remove := range []func() error{
		func() error {
			return eslogstore.Remove(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.ResourceNames.InternalLogStoreSecret)
		},
		clusterRequest.removeKibana,
		func() error {
			return lokistack.RemoveRbac(clusterRequest.Client)
		}} {
		if err := remove(); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Error removing component")
		}
	}
}

func ReconcileForLogFileMetricExporter(lfmeInstance *loggingv1alpha1.LogFileMetricExporter,
	requestClient client.Client,
	er record.EventRecorder,
	clusterID string,
	owner metav1.OwnerReference) (err error) {

	// Make the daemonset along with metric services for Log file metric exporter
	if err := logmetricexporter.Reconcile(lfmeInstance, requestClient, er, owner); err != nil {
		return err
	}

	return nil
}
