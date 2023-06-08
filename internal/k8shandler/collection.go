package k8shandler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/ViaQ/logerr/v2/log/static"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CreateOrUpdateCollection component of the cluster
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCollection() (err error) {
	if !clusterRequest.isManaged() {
		return nil
	}
	cluster := clusterRequest.Cluster
	collectorConfig := ""
	collectorConfHash := ""
	log.V(9).Info("Entering CreateOrUpdateCollection")
	log.V(3).Info("creating collector using", "spec", cluster.Spec.Collection)
	defer func() {
		log.V(9).Info("Leaving CreateOrUpdateCollection")
	}()

	// there is no easier way to check this in golang without writing a helper function
	// TODO: write a helper function to validate Type is a valid option for common setup or tear down
	if cluster.Spec.Collection != nil && cluster.Spec.Collection.Type.IsSupportedCollector() {

		var collectorType = cluster.Spec.Collection.Type

		//TODO: Remove me once fully migrated to new collector naming
		if err = clusterRequest.removeCollector(constants.FluentdName); err != nil {
			log.V(2).Info("Error removing legacy fluentd collector.  ", "err", err)
		}

		if err = clusterRequest.removeCollectorSecretIfOwnedByCLO(); err != nil {
			log.Error(err, "Can't fully clean up old secret created by CLO")
			return
		}

		// LOG-2620: containers violate PodSecurity
		if err = clusterRequest.addSecurityLabelsToNamespace(); err != nil {
			log.Error(err, "Error adding labels to logging Namespace")
			return
		}

		if err = collector.ReconcileServiceAccount(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorServiceAccountName, utils.AsOwner(cluster)); err != nil {
			log.V(9).Error(err, "collector.ReconcileServiceAccount")
			return
		}
		if err = collector.ReconcileRBAC(clusterRequest.Client, cluster.Namespace, constants.CollectorServiceAccountName, utils.AsOwner(cluster)); err != nil {
			log.V(9).Error(err, "collector.ReconcileRBAC")
			return
		}

		if collectorConfig, err = clusterRequest.generateCollectorConfig(); err != nil {
			log.V(9).Error(err, "clusterRequest.generateCollectorConfig")
			return
		}

		log.V(3).Info("Generated collector config", "config", collectorConfig)
		collectorConfHash, err = utils.CalculateMD5Hash(collectorConfig)
		if err != nil {
			log.Error(err, "unable to calculate MD5 hash")
			log.V(9).Error(err, "Returning from unable to calculate MD5 hash")
			return
		}

		instance := clusterRequest.Cluster
		factory := collector.New(collectorConfHash, clusterRequest.ClusterID, *instance.Spec.Collection, clusterRequest.OutputSecrets, clusterRequest.Forwarder.Spec, instance.Name)

		if err := network.ReconcileService(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorName, collector.MetricsPortName, constants.CollectorMetricSecretName, collector.MetricsPort, utils.AsOwner(cluster), factory.CommonLabelInitializer); err != nil {
			log.Error(err, "collector.ReconcileService")
			return err
		}

		if err := metrics.ReconcileServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorName, collector.MetricsPortName, utils.AsOwner(cluster)); err != nil {
			log.Error(err, "collector.ReconcileServiceMonitor")
			return err
		}

		if err := collector.ReconcilePrometheusRule(clusterRequest.EventRecorder, clusterRequest.Client, collectorType, cluster.Namespace, constants.CollectorName, utils.AsOwner(cluster)); err != nil {
			log.V(9).Error(err, "collector.ReconcilePrometheusRule")
		}

		if err = factory.ReconcileCollectorConfig(clusterRequest.EventRecorder, clusterRequest.Client, instance.Namespace, constants.CollectorName, collectorConfig, utils.AsOwner(instance)); err != nil {
			log.Error(err, "collector.ReconcileCollectorConfig")
			return
		}

		if err := collector.ReconcileTrustedCABundleConfigMap(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorTrustedCAName, utils.AsOwner(cluster)); err != nil {
			log.Error(err, "collector.ReconcileTrustedCABundleConfigMap")
			return err
		}

		if err := factory.ReconcileDaemonset(clusterRequest.EventRecorder, clusterRequest.Client, instance.Namespace, constants.CollectorName, utils.AsOwner(instance)); err != nil {
			log.Error(err, "collector.ReconcileDaemonset")
			return err
		}

		// TODO Move me out of here
		if err = clusterRequest.UpdateCollectorStatus(collectorType); err != nil {
			log.V(9).Error(err, "unable to update status for the collector")
		}
	} else {
		if err = clusterRequest.RemoveServiceAccount(constants.CollectorServiceAccountName); err != nil {
			return
		}

		if err = clusterRequest.removeCollector(constants.CollectorName); err != nil {
			return
		}
	}

	return nil
}

// need for smooth upgrade CLO to the 5.4 version, after moving certificates generation to the EO side
// see details: https://issues.redhat.com/browse/LOG-1923
func (clusterRequest *ClusterLoggingRequest) removeCollectorSecretIfOwnedByCLO() (err error) {
	secret, err := clusterRequest.GetSecret(constants.CollectorSecretName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if utils.IsOwnedBy(secret.GetOwnerReferences(), utils.AsOwner(clusterRequest.Cluster)) {
		err = clusterRequest.RemoveSecret(constants.CollectorSecretName)
		if err != nil && !errors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("Can't remove %s secret", constants.CollectorSecretName))
			return err
		}
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeCollector(name string) (err error) {
	log.V(3).Info("Removing collector", "name", name)
	if clusterRequest.isManaged() {

		// https://issues.redhat.com/browse/LOG-3233  Assume if the DS doesn't exist
		// everything is removed
		ds := runtime.NewDaemonSet(clusterRequest.Cluster.Namespace, name)
		key := client.ObjectKeyFromObject(ds)
		if err := clusterRequest.Client.Get(context.TODO(), key, ds); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if err = clusterRequest.RemoveService(name); err != nil {
			return
		}

		metrics.RemoveServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Cluster.Namespace, constants.CollectorName)

		if err = clusterRequest.RemovePrometheusRule(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(name); err != nil {
			return
		}

		caName := fmt.Sprintf("%s-trusted-ca-bundle", name)
		if err = clusterRequest.RemoveConfigMap(caName); err != nil {
			return
		}

		if err = clusterRequest.RemoveDaemonset(name); err != nil {
			return
		}

		// Wait longer than the terminationGracePeriodSeconds
		time.Sleep(12 * time.Second)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) UpdateCollectorStatus(collectorType logging.LogCollectionType) (err error) {
	if collectorType == logging.LogCollectionTypeFluentd {
		return clusterRequest.UpdateFluentdStatus()
	}
	return nil
}

// UpdateFluentdStatus will modify the CL status with the expectation it will be persisted when
// ClusterLogging reconciliation is completed
func (clusterRequest *ClusterLoggingRequest) UpdateFluentdStatus() (err error) {
	fluentdStatus, err := clusterRequest.getFluentdCollectorStatus()
	if err != nil {
		return fmt.Errorf("Failed to get status of the collector: %v", err)
	}
	if !compareFluentdCollectorStatus(fluentdStatus, clusterRequest.Cluster.Status.Collection.Logs.FluentdStatus) {
		clusterRequest.Cluster.Status.Collection.Logs.FluentdStatus = fluentdStatus
	}
	return nil
}

func compareFluentdCollectorStatus(lhs, rhs logging.FluentdCollectorStatus) bool {
	if lhs.DaemonSet != rhs.DaemonSet {
		return false
	}

	if len(lhs.Conditions) != len(rhs.Conditions) {
		return false
	}

	if len(lhs.Conditions) > 0 {
		if !reflect.DeepEqual(lhs.Conditions, rhs.Conditions) {
			return false
		}
	}

	if len(lhs.Nodes) != len(rhs.Nodes) {
		return false
	}

	if len(lhs.Nodes) > 0 {
		if !reflect.DeepEqual(lhs.Nodes, rhs.Nodes) {

			return false
		}
	}

	if len(lhs.Pods) != len(rhs.Pods) {
		return false
	}

	if len(lhs.Pods) > 0 {
		if !reflect.DeepEqual(lhs.Pods, rhs.Pods) {
			return false
		}
	}

	return true
}

func (clusterRequest *ClusterLoggingRequest) addSecurityLabelsToNamespace() error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRequest.Cluster.Namespace,
		},
	}

	key := types.NamespacedName{Name: ns.Name}
	if err := clusterRequest.Client.Get(context.TODO(), key, ns); err != nil {
		return fmt.Errorf("error getting namespace: %w", err)
	}

	if val := ns.Labels[constants.PodSecurityLabelEnforce]; val != constants.PodSecurityLabelValue {
		ns.Labels[constants.PodSecurityLabelEnforce] = constants.PodSecurityLabelValue
		ns.Labels[constants.PodSecurityLabelAudit] = constants.PodSecurityLabelValue
		ns.Labels[constants.PodSecurityLabelWarn] = constants.PodSecurityLabelValue
		ns.Labels[constants.PodSecuritySyncLabel] = "false"

		if err := clusterRequest.Client.Update(context.TODO(), ns); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("error updating namespace: %w", err)
		}
		log.V(1).Info("Successfully added pod security labels", "labels", ns.Labels)
	}

	return nil
}
