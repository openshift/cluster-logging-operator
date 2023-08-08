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
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func Reconcile(cl *logging.ClusterLogging, forwarder *logging.ClusterLogForwarder, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string, resourceNames *factory.ForwarderResourceNames) (err error) {
	log.V(3).Info("Reconciling", "ClusterLogging", cl, "ClusterLogForwarder", forwarder)
	clusterLoggingRequest := ClusterLoggingRequest{
		Cluster:        cl,
		Client:         requestClient,
		Reader:         reader,
		EventRecorder:  r,
		Forwarder:      forwarder,
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
		ResourceNames:  resourceNames,
	}

	// Owner is always the forwarder unless its a "virtual" resource (e.g. legacy CLF)
	// which means CL is not virtual and should have a UID
	if forwarder.UID != "" {
		clusterLoggingRequest.ResourceOwner = utils.AsOwner(forwarder)
	} else {
		clusterLoggingRequest.ResourceOwner = utils.AsOwner(cl)
	}

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

	if clusterLoggingRequest.IsLegacyDeployment() {

		if clusterLoggingRequest.IncludesManagedStorage() {
			// Reconcile Log Store
			if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
				telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
				return fmt.Errorf("unable to create or update logstore for %q: %v", clusterLoggingRequest.Cluster.Name, err)
			}

			// Reconcile Visualization
			if err = clusterLoggingRequest.CreateOrUpdateVisualization(); err != nil {
				telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
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
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		telemetry.Data.CollectorErrorCount.Inc("CollectorErrorCount")
		return fmt.Errorf("unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile metrics Dashboards
	if err = metrics.ReconcileDashboards(clusterLoggingRequest.Client, clusterLoggingRequest.Reader, clusterLoggingRequest.Cluster.Spec.Collection); err != nil {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.Error(err, "Unable to create or update metrics dashboards", "clusterName", clusterLoggingRequest.Cluster.Name)
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

	if updateError := clusterRequest.UpdateCondition(
		logging.CollectorDeadEnd,
		"Collectors are defined but there is no defined LogStore or LogForward destinations",
		"No defined logstore or logforward destination",
		corev1.ConditionTrue,
	); updateError != nil {
		log.Error(updateError, "Unable to update the clusterlogging status", "conditionType", logging.CollectorDeadEnd)
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
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
			return lokistack.RemoveRbac(clusterRequest.Client, func(identifier string) error {
				return RemoveFinalizer(clusterRequest.Client, clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name, identifier)
			})
		}} {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
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
	// Successfully reconciled a LFME, set telemetry to deployed
	telemetry.SetLFMEMetrics(1)
	return nil
}
