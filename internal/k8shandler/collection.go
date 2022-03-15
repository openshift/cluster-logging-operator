package k8shandler

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/daemonsets"
	"path"
	"reflect"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/servicemonitor"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/services"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	exporterPort     = int32(2112)
	exporterPortName = "logfile-metrics"
	metricsPort      = int32(24231)
	metricsPortName  = "metrics"
	prometheusCAFile = "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt"

	AnnotationDebugOutput = "logging.openshift.io/debug-output"
)

//CreateOrUpdateCollection component of the cluster
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCollection() (err error) {
	cluster := clusterRequest.Cluster
	collectorConfig := ""
	collectorConfHash := ""
	log.V(9).Info("Entering CreateOrUpdateCollection")
	defer func() {
		log.V(9).Info("Leaving CreateOrUpdateCollection")
	}()

	var collectorServiceAccount *core.ServiceAccount

	// there is no easier way to check this in golang without writing a helper function
	// TODO: write a helper function to validate Type is a valid option for common setup or tear down
	if cluster.Spec.Collection != nil &&
		(cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeFluentd ||
			cluster.Spec.Collection.Logs.Type == logging.LogCollectionTypeVector) {

		var collectorType = cluster.Spec.Collection.Logs.Type

		//TODO: Remove me once fully migrated to new collector naming
		if err = clusterRequest.removeCollector(constants.FluentdName); err != nil {
			log.V(2).Info("Error removing legacy fluentd collector.  ", "err", err)
		}
		enabled, found := clusterRequest.Cluster.Annotations[common.PreviewVectorCollector]
		if collectorType == logging.LogCollectionTypeVector && (!found || enabled != "enabled") {
			err = errors.NewBadRequest(fmt.Sprintf("Vector as collector not enabled via annotation on ClusterLogging %s", common.PreviewVectorCollector))
			log.V(9).Error(err, "Vector as collector not enabled via annotation on ClusterLogging")
			return clusterRequest.UpdateCondition(
				logging.CollectorDeadEnd,
				"Add annotation \"logging.openshift.io/preview-vector-collector: enabled\" to ClusterLogging CR",
				"Vector as collector not enabled via annotation on ClusterLogging",
				corev1.ConditionTrue,
			)
		}

		if err = clusterRequest.removeCollectorSecretIfOwnedByCLO(); err != nil {
			log.Error(err, "Can't fully clean up old secret created by CLO")
			return
		}

		if collectorServiceAccount, err = clusterRequest.createOrUpdateCollectorServiceAccount(); err != nil {
			log.V(9).Error(err, "clusterRequest.createOrUpdateCollectorServiceAccount")
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
		if err = clusterRequest.reconcileCollectorService(); err != nil {
			log.V(9).Error(err, "clusterRequest.reconcileCollectorService")
			return
		}

		if err = clusterRequest.reconcileCollectorServiceMonitor(); err != nil {
			log.V(9).Error(err, "clusterRequest.reconcileCollectorServiceMonitor")
			return
		}

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
	if clusterRequest.isManaged() {

		if err = clusterRequest.RemoveService(name); err != nil {
			return
		}

		if err = clusterRequest.RemoveServiceMonitor(name); err != nil {
			return
		}

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
				Port:       metricsPort,
				TargetPort: intstr.FromString(metricsPortName),
				Name:       metricsPortName,
			},
			{
				Port:       exporterPort,
				TargetPort: intstr.FromString(exporterPortName),
				Name:       exporterPortName,
			},
		},
	)

	desired.Annotations = map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": constants.CollectorMetricSecretName,
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
		log.V(3).Error(retryErr, "Reconcile Service retry error")
		return retryErr
	}
	return err
}

func (clusterRequest *ClusterLoggingRequest) reconcileCollectorServiceMonitor() error {

	cluster := clusterRequest.Cluster

	desired := NewServiceMonitor(constants.CollectorName, cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   metricsPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", constants.CollectorName, cluster.Namespace),
		},
	}
	logMetricExporterEndpoint := monitoringv1.Endpoint{
		Port:   exporterPortName,
		Path:   "/metrics",
		Scheme: "https",
		TLSConfig: &monitoringv1.TLSConfig{
			CAFile:     prometheusCAFile,
			ServerName: fmt.Sprintf("%s.%s.svc", constants.CollectorName, cluster.Namespace),
		},
	}

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"logging-infra": "support",
		},
	}

	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-collector",
		Endpoints: []monitoringv1.Endpoint{endpoint, logMetricExporterEndpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(desired, utils.AsOwner(cluster))

	err := clusterRequest.Create(desired)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating the collector ServiceMonitor: %v", err)
		}
		current := &monitoringv1.ServiceMonitor{}
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(desired.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %q service for %q: %v", current.Name, clusterRequest.Cluster.Name, err)
			}
			if servicemonitor.AreSame(current, desired) {
				log.V(3).Info("ServiceMonitor are the same skipping update")
				return nil
			}
			current.Labels = desired.Labels
			current.Spec = desired.Spec
			current.Annotations = desired.Annotations

			return clusterRequest.Update(current)
		})
		log.V(3).Error(retryErr, "Reconcile ServiceMonitor retry error")
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
	var err error = nil
	if collectorType == logging.LogCollectionTypeFluentd {
		collectorConfigMap := NewConfigMap(
			constants.CollectorName,
			clusterRequest.Cluster.Namespace,
			map[string]string{
				"fluent.conf":         collectorConfig,
				"run.sh":              fluentd.RunScript,
				"cleanInValidJson.rb": fluentd.CleanInValidJson,
			},
		)
		err = clusterRequest.createConfigMap(collectorConfigMap)

	} else if collectorType == logging.LogCollectionTypeVector {
		var secrets = map[string][]byte{}
		_ = Syncronize(func() error {
			secrets = map[string][]byte{}
			// Ignore errors, these files are optional depending on security settings.
			secrets["vector.toml"] = []byte(collectorConfig)
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
	desired := collector.NewDaemonSet(instance.Namespace, configHash, caTrustHash, collectorType, caTrustBundle, *instance.Spec.Collection, clusterRequest.ForwarderSpec, clusterRequest.OutputSecrets)
	utils.AddOwnerRefToObject(desired, utils.AsOwner(instance))

	current := &apps.DaemonSet{}
	err = clusterRequest.Get(desired.Name, current)
	if errors.IsNotFound(err) {
		if err = clusterRequest.Create(desired); err != nil {
			return fmt.Errorf("Failure creating collector Daemonset %v", err)
		}
		return nil
	}
	if daemonsets.AreSame(current, desired) {
		return nil
	}
	//With this PR: https://github.com/kubernetes-sigs/controller-runtime/pull/919
	//we have got new behaviour: Reset resource version if fake client Create call failed.
	//So if object already exist version will be reset, going to get before try to create.
	desired.ResourceVersion = current.GetResourceVersion()
	current.Spec = desired.Spec
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err = clusterRequest.Update(desired); err != nil {
			return err
		}
		return nil
	})
	if retryErr != nil {
		return retryErr
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
