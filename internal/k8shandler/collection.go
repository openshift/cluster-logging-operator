package k8shandler

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"path"
	"reflect"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"

	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/tls"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/services"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationDebugOutput = "logging.openshift.io/debug-output"
)

//CreateOrUpdateCollection component of the cluster
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCollection() (err error) {
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
		if err = clusterRequest.reconcileCollectorService(); err != nil {
			log.V(9).Error(err, "clusterRequest.reconcileCollectorService")
			return
		}

		collector.ReconcileServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, cluster.Namespace, constants.CollectorName, utils.AsOwner(cluster))

		if err = clusterRequest.createOrUpdateCollectorPrometheusRule(); err != nil {
			log.V(9).Error(err, "unable to create or update fluentd prometheus rule")
		}

		if err = clusterRequest.createOrUpdateCollectorConfig(collectorType, collectorConfig); err != nil {
			log.V(9).Error(err, "clusterRequest.createOrUpdateCollectorConfig")
			return
		}

		if err = clusterRequest.reconcileCollectorDaemonset(collectorType, collectorConfHash); err != nil {
			log.V(9).Error(err, "clusterRequest.reconcileCollectorDaemonset")
			return
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

func (clusterRequest *ClusterLoggingRequest) reconcileCollectorService() error {
	desired := factory.NewService(
		constants.CollectorName,
		clusterRequest.Cluster.Namespace,
		constants.CollectorName,
		[]v1.ServicePort{
			{
				Port:       collector.MetricsPort,
				TargetPort: intstr.FromString(collector.MetricsPortName),
				Name:       collector.MetricsPortName,
			},
			{
				Port:       collector.ExporterPort,
				TargetPort: intstr.FromString(collector.ExporterPortName),
				Name:       collector.ExporterPortName,
			},
		},
	)

	desired.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: constants.CollectorMetricSecretName,
	}

	utils.AddOwnerRefToObject(desired, utils.AsOwner(clusterRequest.Cluster))
	err := clusterRequest.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating the collector service: %v", err)
		}
		current := &v1.Service{}
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %q service for %q: %v", current.Name, clusterRequest.Cluster.Name, err)
			}
			if services.AreSame(current, desired) {
				log.V(3).Info("Services are the same skipping update")
				return nil
			}
			//Explicitly copying because services are immutable
			current.Labels = desired.Labels
			current.Spec.Selector = desired.Spec.Selector
			current.Spec.Ports = desired.Spec.Ports
			return clusterRequest.Update(current)
		})
		if retryErr != nil {
			log.V(3).Error(retryErr, "Reconcile Service retry error")
		}
		return retryErr
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorPrometheusRule() error {
	ctx := context.TODO()
	cluster := clusterRequest.Cluster

	rule := NewPrometheusRule(constants.CollectorName, cluster.Namespace)

	spec, err := NewPrometheusRuleSpecFrom(path.Join(utils.GetShareDir(), fluentdAlertsFile))
	if err != nil {
		return fmt.Errorf("failure creating the collector PrometheusRule: %w", err)
	}

	rule.Spec = *spec

	utils.AddOwnerRefToObject(rule, utils.AsOwner(cluster))

	err = clusterRequest.Create(rule)
	if err == nil {
		return nil
	}
	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failure creating the collector PrometheusRule: %w", err)
	}

	current := &monitoringv1.PrometheusRule{}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = clusterRequest.Client.Get(ctx, types.NamespacedName{Name: rule.Name, Namespace: rule.Namespace}, current)
		if err != nil {
			log.V(2).Info("could not get prometheus rule", rule.Name, err)
			return err
		}
		current.Spec = rule.Spec
		if err = clusterRequest.Client.Update(ctx, current); err != nil {
			return err
		}
		log.V(3).Info("updated prometheus rules")
		return nil
	})
}

func (clusterRequest *ClusterLoggingRequest) createOrUpdateCollectorConfig(collectorType logging.LogCollectionType, collectorConfig string) error {
	log.V(3).Info("Updating ConfigMap and Secrets")
	opensslConf := tls.OpenSSLConf(clusterRequest.Client)
	var err error = nil
	if collectorType == logging.LogCollectionTypeFluentd {
		collectorConfigMap := NewConfigMap(
			constants.CollectorName,
			clusterRequest.Cluster.Namespace,
			map[string]string{
				"fluent.conf":         collectorConfig,
				"run.sh":              fluentd.RunScript,
				"cleanInValidJson.rb": fluentd.CleanInValidJson,
				"openssl.cnf":         opensslConf,
			},
		)
		err = clusterRequest.createConfigMap(collectorConfigMap)

	} else if collectorType == logging.LogCollectionTypeVector {
		var secrets = map[string][]byte{}
		_ = Syncronize(func() error {
			secrets = map[string][]byte{}
			// Ignore errors, these files are optional depending on security settings.
			secrets["vector.toml"] = []byte(collectorConfig)
			secrets["openssl.cnf"] = []byte(opensslConf)
			return nil
		})
		err = clusterRequest.createSecret(constants.CollectorConfigSecretName, clusterRequest.Cluster.Namespace, secrets)
	}
	if err != nil {
		return err
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) createConfigMap(collectorConfigMap *corev1.ConfigMap) error {
	utils.AddOwnerRefToObject(collectorConfigMap, utils.AsOwner(clusterRequest.Cluster))

	err := clusterRequest.Create(collectorConfigMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing collector configmap: %v", err)
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &v1.ConfigMap{}
		if err = clusterRequest.Get(collectorConfigMap.Name, current); err != nil {
			if errors.IsNotFound(err) {
				log.V(2).Info("Returning nil. The configmap was not found even though create previously failed.  Was it culled?", "configmap name", collectorConfigMap.Name)
				return nil
			}
			return fmt.Errorf("Failed to get %v configmap for %q: %v", collectorConfigMap.Name, clusterRequest.Cluster.Name, err)
		}
		if reflect.DeepEqual(collectorConfigMap.Data, current.Data) {
			return nil
		}
		current.Data = collectorConfigMap.Data
		return clusterRequest.Update(current)
	})
	return retryErr
}

func (clusterRequest *ClusterLoggingRequest) createSecret(secretName, namespace string, secrets map[string][]byte) error {
	collectorSecret := NewSecret(secretName, namespace, secrets)
	utils.AddOwnerRefToObject(collectorSecret, utils.AsOwner(clusterRequest.Cluster))
	err := clusterRequest.CreateOrUpdateSecret(collectorSecret)
	return err
}

func (clusterRequest *ClusterLoggingRequest) reconcileCollectorDaemonset(collectorType logging.LogCollectionType, configHash string) (err error) {
	if !clusterRequest.isManaged() {
		return nil
	}
	var caTrustBundle *v1.ConfigMap
	// Create or update cluster proxy trusted CA bundle.
	caTrustBundle, err = clusterRequest.createOrGetTrustedCABundleConfigMap(constants.CollectorTrustedCAName)
	if err != nil {
		return
	}
	caTrustHash, err := calcTrustedCAHashValue(caTrustBundle)
	if err != nil || caTrustHash == "" {
		log.V(1).Info("Cluster wide proxy may not be configured. ConfigMap does not contain expected key or does not contain ca bundle", "configmapName", constants.CollectorTrustedCAName, "key", constants.TrustedCABundleKey, "err", err)
	}

	instance := clusterRequest.Cluster
	desired := collector.NewDaemonSet(instance.Namespace, configHash, caTrustHash, clusterRequest.ClusterID, collectorType, caTrustBundle, *instance.Spec.Collection, clusterRequest.ForwarderSpec, clusterRequest.OutputSecrets)
	utils.AddOwnerRefToObject(desired, utils.AsOwner(instance))

	return reconcile.DaemonSet(clusterRequest.EventRecorder, clusterRequest.Client, desired)
}

func (clusterRequest *ClusterLoggingRequest) UpdateCollectorStatus(collectorType logging.LogCollectionType) (err error) {
	if collectorType == logging.LogCollectionTypeFluentd {
		return clusterRequest.UpdateFluentdStatus()
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) UpdateFluentdStatus() (err error) {

	cluster := clusterRequest.Cluster

	fluentdStatus, err := clusterRequest.getFluentdCollectorStatus()
	if err != nil {
		return fmt.Errorf("Failed to get status of the collector: %v", err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if !compareFluentdCollectorStatus(fluentdStatus, cluster.Status.Collection.Logs.FluentdStatus) {
			cluster.Status.Collection.Logs.FluentdStatus = fluentdStatus
			return clusterRequest.UpdateStatus(cluster)
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

	subject := NewSubject(
		"ServiceAccount",
		constants.CollectorServiceAccountName,
	)
	subject.APIGroup = ""

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
