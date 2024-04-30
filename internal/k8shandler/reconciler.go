package k8shandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder"

	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/metrics"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	"github.com/openshift/cluster-logging-operator/internal/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"k8s.io/apimachinery/pkg/types"
)

func Reconcile(cl *logging.ClusterLogging, requestClient client.Client, reader client.Reader, r record.EventRecorder, clusterVersion, clusterID string) (instance *logging.ClusterLogging, err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Cluster:        cl,
		Client:         requestClient,
		Reader:         reader,
		EventRecorder:  r,
		ClusterVersion: clusterVersion,
		ClusterID:      clusterID,
	}

	if instance, err = clusterLoggingRequest.getClusterLogging(false); err != nil {
		return nil, err
	}
	clusterLoggingRequest.Cluster = instance

	if instance.GetDeletionTimestamp() != nil {
		// ClusterLogging is being deleted, remove resources that can not be garbage-collected.
		if err := clusterLoggingRequest.removeLokiStackRbac(); err != nil {
			log.Error(err, "Error removing RBAC for accessing LokiStack.")
		}
	}

	if !clusterLoggingRequest.isManaged() {
		return clusterLoggingRequest.Cluster, nil
	}
	forwarder, extras := clusterLoggingRequest.getLogForwarder()
	if forwarder != nil {
		if err := clusterlogforwarder.Validate(*forwarder); err != nil {
			return nil, err
		}
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	} else if !clusterLoggingRequest.IncludesManagedStorage() {
		// No clf and no logStore so remove the collector https://issues.redhat.com/browse/LOG-2703
		removeCollectorAndUpdate(clusterLoggingRequest)
		return clusterLoggingRequest.Cluster, nil
	}

	if clusterLoggingRequest.IncludesManagedStorage() {
		// Reconcile Log Store
		if err = clusterLoggingRequest.CreateOrUpdateLogStore(); err != nil {
			return clusterLoggingRequest.Cluster, fmt.Errorf("unable to create or update logstore for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

		// Reconcile Visualization
		if err = clusterLoggingRequest.CreateOrUpdateVisualization(); err != nil {
			return clusterLoggingRequest.Cluster, fmt.Errorf("unable to create or update visualization for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		}

	} else {
		removeManagedStorage(clusterLoggingRequest)
	}

	// Remove Curator
	if err := clusterLoggingRequest.removeCurator(); err != nil {
		log.V(0).Error(err, "Error removing curator component")
	}
	clusterLoggingRequest.Cluster.Status.Conditions.SetCondition(status.Condition{
		Type:    "CuratorRemoved",
		Status:  corev1.ConditionTrue,
		Reason:  "ResourceDeprecated",
		Message: "curator is deprecated in favor of defining retention policy",
	})

	// Reconcile Collection
	if err = clusterLoggingRequest.CreateOrUpdateCollection(extras); err != nil {
		telemetry.IncreaseCollectorErrors()
		return clusterLoggingRequest.Cluster, fmt.Errorf("unable to create or update collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
	}

	// Reconcile metrics Dashboards
	if err = metrics.ReconcileDashboards(clusterLoggingRequest.Client, reader, clusterLoggingRequest.Cluster.Spec.Collection); err != nil {
		log.Error(err, "Unable to create or update metrics dashboards", "clusterName", clusterLoggingRequest.Cluster.Name)
	}

	telemetry.UpdateDefaultForwarderInfo(clusterLoggingRequest.ForwarderRequest)

	return clusterLoggingRequest.Cluster, nil
}

func removeCollectorAndUpdate(clusterRequest ClusterLoggingRequest) {
	log.V(3).Info("forwarder not found and logStore not found so removing collector")
	if err := clusterRequest.removeCollector(constants.CollectorName); err != nil {
		log.Error(err, "Error removing collector")
	}

	if updateError := clusterRequest.UpdateCondition(
		logging.CollectorDeadEnd,
		"Collectors are defined but there is no defined LogStore or LogForward destinations",
		"No defined logstore or logforward destination",
		corev1.ConditionTrue,
	); updateError != nil {
		log.Error(updateError, "Unable to update the clusterlogging status", "conditionType", logging.CollectorDeadEnd)
	}
}

func removeManagedStorage(clusterRequest ClusterLoggingRequest) {
	log.V(1).Info("Removing managed store components...")
	for _, remove := range []func() error{clusterRequest.removeElasticsearch, clusterRequest.removeKibana, clusterRequest.removeLokiStackRbac} {
		if err := remove(); err != nil {
			log.Error(err, "Error removing component")
		}
	}
}

func ReconcileForClusterLogForwarder(forwarder *logging.ClusterLogForwarder, requestClient client.Client, er record.EventRecorder, clusterID string) (err error) {
	clusterLoggingRequest := ClusterLoggingRequest{
		Client:        requestClient,
		EventRecorder: er,
		ClusterID:     clusterID,
	}
	if forwarder != nil {
		clusterLoggingRequest.ForwarderRequest = forwarder
		clusterLoggingRequest.ForwarderSpec = forwarder.Spec
	}

	var clusterLogging *logging.ClusterLogging
	if clusterLogging, err = clusterLoggingRequest.getClusterLogging(false); err != nil {
		return err
	}
	if clusterLogging == nil {
		return nil
	}

	extras := map[string]bool{}
	clusterLoggingRequest.ForwarderSpec, extras = migrations.MigrateClusterLogForwarderSpec(forwarder.Spec, clusterLogging.Spec.LogStore, extras)
	clusterLoggingRequest.Cluster = clusterLogging

	if clusterLogging.Spec.ManagementState == logging.ManagementStateUnmanaged {
		return nil
	}

	// Reconcile Collection
	err = clusterLoggingRequest.CreateOrUpdateCollection(extras)
	forwarder.Status = clusterLoggingRequest.ForwarderRequest.Status

	if err != nil {
		msg := fmt.Sprintf("Unable to reconcile collection for %q: %v", clusterLoggingRequest.Cluster.Name, err)
		log.Error(err, msg)
		return errors.New(msg)
	}

	telemetry.UpdateDefaultForwarderInfo(clusterLoggingRequest.ForwarderRequest)

	return nil
}

func (clusterRequest *ClusterLoggingRequest) getClusterLogging(skipMigrations bool) (*logging.ClusterLogging, error) {
	clusterLoggingNamespacedName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.WatchNamespace}
	clusterLogging := &logging.ClusterLogging{}

	if err := clusterRequest.Client.Get(context.TODO(), clusterLoggingNamespacedName, clusterLogging); err != nil {
		return nil, err
	}

	// Do not modify cached copy
	clusterLogging = clusterLogging.DeepCopy()

	if skipMigrations {
		return clusterLogging, nil
	}

	// TODO Drop migration upon introduction of v2
	clusterLogging.Spec = migrations.MigrateCollectionSpec(clusterLogging.Spec)

	return clusterLogging, nil
}

func (clusterRequest *ClusterLoggingRequest) getLogForwarder() (*logging.ClusterLogForwarder, map[string]bool) {
	nsname := types.NamespacedName{Name: constants.SingletonName, Namespace: clusterRequest.Cluster.Namespace}
	forwarder := runtime.NewClusterLogForwarder(clusterRequest.Cluster.Namespace, clusterRequest.Cluster.Name)
	if err := clusterRequest.Client.Get(context.TODO(), nsname, forwarder); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "forwarder", nsname)
		}
		if !clusterRequest.IncludesManagedStorage() {
			return nil, map[string]bool{}
		}
		forwarder.Spec = logging.ClusterLogForwarderSpec{}
	}
	extras := map[string]bool{}
	forwarder.Spec, extras = migrations.MigrateClusterLogForwarderSpec(forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras)
	return forwarder, extras
}
