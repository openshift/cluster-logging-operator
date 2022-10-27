package k8shandler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"

	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationDebugOutput = "logging.openshift.io/debug-output"
)

//CreateOrUpdateCollection component of the cluster
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

	var collectorServiceAccount *core.ServiceAccount

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

		if collectorServiceAccount, err = clusterRequest.createOrUpdateCollectorServiceAccount(); err != nil {
			log.V(9).Error(err, "clusterRequest.createOrUpdateCollectorServiceAccount")
			return
		} else {
			if err = clusterRequest.createOrUpdateCollectorTokenSecret(); err != nil {
				log.V(9).Error(err, "clusterRequest.createOrUpdateCollectorTokenSecret")
				return
			}
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

		if err := collector.ReconcileService(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorName, utils.AsOwner(cluster)); err != nil {
			log.Error(err, "collector.ReconcileService")
			return err
		}

		if err := collector.ReconcileServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorName, utils.AsOwner(cluster)); err != nil {
			log.Error(err, "collector.ReconcileServiceMonitor")
			return err
		}

		if err := collector.ReconcilePrometheusRule(clusterRequest.EventRecorder, clusterRequest.Client, collectorType, cluster.Namespace, constants.CollectorName, utils.AsOwner(cluster)); err != nil {
			log.V(9).Error(err, "collector.ReconcilePrometheusRule")
		}

		instance := clusterRequest.Cluster
		factory := collector.New(collectorConfHash, clusterRequest.ClusterID, *instance.Spec.Collection, clusterRequest.OutputSecrets, clusterRequest.ForwarderSpec)

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

		if err = clusterRequest.UpdateCollectorStatus(collectorType); err != nil {
			log.V(9).Error(err, "unable to update status for the collector")
		}

		if collectorServiceAccount != nil {

			// remove our finalizer from the list and update it.
			collectorServiceAccount.ObjectMeta.Finalizers = utils.RemoveString(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
			if err = clusterRequest.Update(collectorServiceAccount); err != nil {
				log.Info("Unable to update the collector serviceaccount finalizers", "collectorServiceAccount.Name", collectorServiceAccount.Name)
				log.V(9).Error(err, "Unable to update the collector serviceaccount finalizers")
				return nil
			}
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

		if err = clusterRequest.RemoveService(name); err != nil {
			return
		}

		collector.RemoveServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Cluster.Namespace, constants.CollectorName)

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

func (clusterRequest *ClusterLoggingRequest) UpdateFluentdStatus() (err error) {
	fluentdStatus, err := clusterRequest.getFluentdCollectorStatus()
	if err != nil {
		return fmt.Errorf("Failed to get status of the collector: %v", err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance, err := clusterRequest.getClusterLogging(true)
		if err != nil {
			return err
		}

		if !compareFluentdCollectorStatus(fluentdStatus, instance.Status.Collection.Logs.FluentdStatus) {
			instance.Status.Collection.Logs.FluentdStatus = fluentdStatus
			return clusterRequest.UpdateStatus(instance)
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging collector status: %v", retryErr)
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

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorServiceAccount() (*core.ServiceAccount, error) {
	collectorServiceAccount := runtime.NewServiceAccount(clusterRequest.Cluster.Namespace, constants.CollectorServiceAccountName)

	utils.AddOwnerRefToObject(collectorServiceAccount, utils.AsOwner(clusterRequest.Cluster))

	delfinalizer := false
	if collectorServiceAccount.ObjectMeta.DeletionTimestamp.IsZero() {
		// This object is not being deleted.
		if !utils.ContainsString(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents) {
			collectorServiceAccount.ObjectMeta.Finalizers = append(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
		}
		err := clusterRequest.Create(collectorServiceAccount)
		if err != nil && !errors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("Failure creating Log collector service account: %v", err)
		}
	} else if utils.ContainsString(collectorServiceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents) {
		// This object is being deleted.
		// our finalizer is present, so lets handle any dependency
		delfinalizer = true
	}

	// If needed create a custom SecurityContextConstraints object for the log collector to run with
	scc := NewSCC(LogCollectorSCCName)
	err := clusterRequest.Create(scc)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Failure creating Log collector SecurityContextConstraints: %v", err)
	}

	if delfinalizer {
		return collectorServiceAccount, nil
	} else {
		return nil, nil
	}
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

		if err := clusterRequest.Client.Update(context.TODO(), ns); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("error updating namespace: %w", err)
		}
		log.V(1).Info("Successfully added pod security labels", "namespace.Labels", ns.Labels)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorTokenSecret() error {
	serviceAccount := &corev1.ServiceAccount{}
	if err := clusterRequest.Get(constants.CollectorServiceAccountName, serviceAccount); err != nil {
		return fmt.Errorf("failed to get ServiceAccount: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterRequest.Cluster.Namespace,
			Name:      constants.LogCollectorToken,
			Annotations: map[string]string{
				corev1.ServiceAccountNameKey: serviceAccount.Name,
				corev1.ServiceAccountUIDKey:  string(serviceAccount.UID),
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))

	if err := clusterRequest.Get(constants.LogCollectorToken, secret); err == nil {
		accountName := secret.Annotations[corev1.ServiceAccountNameKey]
		accountUID := secret.Annotations[corev1.ServiceAccountUIDKey]
		if accountName != serviceAccount.Name || accountUID != string(serviceAccount.UID) {
			// Delete secret, so that we can create a new one next loop
			if err := clusterRequest.Delete(secret); err != nil {
				return err
			}

			return fmt.Errorf("deleted stale secret: %s", constants.LogCollectorToken)
		}

		// Existing secret is up-to-date
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get logcollector token secret: %w", err)
	}

	if err := clusterRequest.Create(secret); err != nil {
		return fmt.Errorf("failed to create logcollector token secret: %w", err)
	}

	return nil
}
