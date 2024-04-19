package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	logmetricexporter "github.com/openshift/cluster-logging-operator/internal/metrics/logfilemetricexporter"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func Reconcile(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string, resourceNames *factory.ForwarderResourceNames) (err error) {
	log.V(3).Info("Reconciling", "ClusterLogging", cl, "ClusterLogForwarder", forwarder)
	clusterLoggingRequest := NewClusterLoggingRequest(cl, forwarder, requestClient, reader, r, clusterVersion, clusterID, resourceNames)

	telemetry.CancelMetrics()
	defer telemetry.UpdateMetrics()

	if !clusterLoggingRequest.isManaged() {
		// if cluster is set to unmanaged then set managedStatus as 0
		telemetry.Data.CLInfo.Set("managedStatus", constants.UnManagedStatus)
		return nil
	}
	// CL is managed by default set it as 1
	telemetry.Data.CLInfo.Set("managedStatus", constants.ManagedStatus)
	telemetry.UpdateInfofromCL(*clusterLoggingRequest.Cluster)

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
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		telemetry.Data.CollectorErrorCount.Inc("CollectorErrorCount")
		return fmt.Errorf("unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Clean up any stale http input services
	if err = clusterLoggingRequest.RemoveInputServices([]metav1.OwnerReference{utils.AsOwner(forwarder)}, false); err != nil {
		return fmt.Errorf("error removing stale http input services. %v", err)
	}

	//if there is no early exit from reconciler then new CL spec is applied successfully hence healthStatus is set to true or 1
	telemetry.Data.CLInfo.Set("healthStatus", constants.HealthyStatus)
	telemetry.UpdateInfofromCLF(*clusterLoggingRequest.Forwarder)
	return nil
}

func removeCollectorAndUpdate(clusterRequest ClusterLoggingRequest) {
	log.V(3).Info("forwarder not found and logStore not found so removing collector")
	if err := clusterRequest.removeCollector(); err != nil {
		log.Error(err, "Error removing collector")
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
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
	// Successfully reconciled a LFME, set telemetry to deployed
	telemetry.SetLFMEMetrics(1)
	return nil
}
