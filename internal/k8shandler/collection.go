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

	// Default to vector collector type
	collectionSpec := &logging.CollectionSpec{
		Type: logging.LogCollectionTypeVector,
	}
	log.V(9).Info("Entering CreateOrUpdateCollection")
	defer func() {
		log.V(9).Info("Leaving CreateOrUpdateCollection")
	}()

	if cluster != nil && clusterRequest.Forwarder.Name == constants.SingletonName {
		if cluster.Spec.Collection != nil && cluster.Spec.Collection.Type.IsSupportedCollector() {
			// Change collector type dependent on clusterLogging
			collectionSpec = cluster.Spec.Collection
		} else {
			if err = clusterRequest.RemoveServiceAccount(); err != nil {
				return
			}
			if err = clusterRequest.removeCollector(); err != nil {
				return
			}
		}
	}

	// LOG-2620: containers violate PodSecurity
	if err = clusterRequest.addSecurityLabelsToNamespace(); err != nil {
		log.Error(err, "Error adding labels to logging Namespace")
		return
	}

	if err = collector.ReconcileServiceAccount(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames, clusterRequest.ResourceOwner); err != nil {
		log.V(9).Error(err, "collector.ReconcileServiceAccount")
		return
	}

	// This also reconciles the ServiceAccount role and role bindings for the SCC
	if err = collector.ReconcileRBAC(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames, clusterRequest.ResourceOwner); err != nil {
		log.V(9).Error(err, "collector.ReconcileRBAC")
		return
	}

	// Set the output secrets if any
	clusterRequest.SetOutputSecrets()
	tokenSecret, err := clusterRequest.GetLogCollectorServiceAccountTokenSecret()
	if err == nil {
		saTokenSecretName := clusterRequest.ResourceNames.ServiceAccountTokenSecret
		clusterRequest.OutputSecrets[saTokenSecretName] = tokenSecret
	}

	if collectorConfig, err = clusterRequest.generateCollectorConfig(); err != nil {
		log.V(9).Error(err, "clusterRequest.generateCollectorConfig")
		return err
	}

	log.V(3).Info("Generated collector config", "config", collectorConfig)
	collectorConfHash, err = utils.CalculateMD5Hash(collectorConfig)
	if err != nil {
		log.Error(err, "unable to calculate MD5 hash")
		log.V(9).Error(err, "Returning from unable to calculate MD5 hash")
		return
	}

	factory := collector.New(collectorConfHash, clusterRequest.ClusterID, *collectionSpec, clusterRequest.OutputSecrets, clusterRequest.Forwarder.Spec, clusterRequest.Forwarder.Name, clusterRequest.ResourceNames)

	if err := network.ReconcileService(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CommonName, collector.MetricsPortName, clusterRequest.ResourceNames.SecretMetrics, collector.MetricsPort, clusterRequest.ResourceOwner, factory.CommonLabelInitializer); err != nil {
		log.Error(err, "collector.ReconcileService")
		return err
	}

	if err := metrics.ReconcileServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CommonName, collector.MetricsPortName, clusterRequest.ResourceOwner); err != nil {
		log.Error(err, "collector.ReconcileServiceMonitor")
		return err
	}
	if err := collector.ReconcilePrometheusRule(clusterRequest.EventRecorder, clusterRequest.Client, collectionSpec.Type, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CommonName, clusterRequest.ResourceOwner); err != nil {
		log.V(9).Error(err, "collector.ReconcilePrometheusRule")
	}

	if err = factory.ReconcileCollectorConfig(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, collectorConfig, clusterRequest.ResourceOwner); err != nil {
		log.Error(err, "collector.ReconcileCollectorConfig")
		return
	}

	if err := collector.ReconcileTrustedCABundleConfigMap(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CaTrustBundle, clusterRequest.ResourceOwner); err != nil {
		log.Error(err, "collector.ReconcileTrustedCABundleConfigMap")
		return err
	}
	if err := factory.ReconcileDaemonset(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceOwner); err != nil {
		log.Error(err, "collector.ReconcileDaemonset")
		return err
	}

	if err = clusterRequest.UpdateCollectorStatus(collectionSpec.Type); err != nil {
		log.V(9).Error(err, "unable to update status for the collector")
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) removeCollector() (err error) {
	commonName := clusterRequest.ResourceNames.CommonName
	log.V(3).Info("Removing collector", "name", commonName)
	if clusterRequest.isManaged() {

		// https://issues.redhat.com/browse/LOG-3233  Assume if the DS doesn't exist
		// everything is removed
		ds := runtime.NewDaemonSet(clusterRequest.Forwarder.Namespace, commonName)
		key := client.ObjectKeyFromObject(ds)
		if err := clusterRequest.Client.Get(context.TODO(), key, ds); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if err = clusterRequest.RemoveService(commonName); err != nil {
			return
		}

		metrics.RemoveServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, commonName)

		if err = clusterRequest.RemovePrometheusRule(commonName); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(clusterRequest.ResourceNames.ConfigMap); err != nil {
			return
		}

		if err = clusterRequest.RemoveConfigMap(clusterRequest.ResourceNames.CaTrustBundle); err != nil {
			return
		}

		if err = clusterRequest.RemoveDaemonset(commonName); err != nil {
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
			Name: clusterRequest.Forwarder.Namespace,
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
