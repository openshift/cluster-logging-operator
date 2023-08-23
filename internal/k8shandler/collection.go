package k8shandler

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/ViaQ/logerr/v2/log/static"

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
	log.V(9).Info("Entering CreateOrUpdateCollection")
	defer func() {
		log.V(9).Info("Leaving CreateOrUpdateCollection")
	}()

	if !clusterRequest.isManaged() {
		return nil
	}
	collectorConfig := ""
	collectorConfHash := ""

	// LOG-2620: containers violate PodSecurity
	if err = clusterRequest.addSecurityLabelsToNamespace(); err != nil {
		log.Error(err, "Error adding labels to logging Namespace")
		return
	}

	// Remove legacy SecurityContextConstraint named `log-collector-scc` before reconciling a new one
	if err = auth.RemoveSecurityContextConstraint(clusterRequest.Client, "log-collector-scc"); err != nil {
		return
	}

	if err = reconcile.SecurityContextConstraints(clusterRequest.Client, auth.NewSCC()); err != nil {
		log.V(9).Error(err, "reconcile.SecurityContextConstraints")
		return err
	}

	if err = auth.ReconcileServiceAccount(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames, clusterRequest.ResourceOwner); err != nil {
		log.V(9).Error(err, "collector.ReconcileServiceAccount")
		return
	}

	// This also reconciles the ServiceAccount role and role bindings for the SCC
	if err = auth.ReconcileRBAC(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames, clusterRequest.ResourceOwner); err != nil {
		log.V(9).Error(err, "collector.ReconcileRBAC")
		return
	}

	// Set the output secrets if any
	clusterRequest.SetOutputSecrets()

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

	factory := collector.New(collectorConfHash, clusterRequest.ClusterID, *clusterRequest.Cluster.Spec.Collection, clusterRequest.OutputSecrets, clusterRequest.Forwarder.Spec, clusterRequest.ResourceNames.CommonName, clusterRequest.ResourceNames)

	if err := network.ReconcileService(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, clusterRequest.ResourceNames.SecretMetrics, collector.MetricsPort, clusterRequest.ResourceOwner, factory.CommonLabelInitializer); err != nil {
		log.Error(err, "collector.ReconcileService")
		return err
	}

	if err := metrics.ReconcileServiceMonitor(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, clusterRequest.ResourceOwner); err != nil {
		log.Error(err, "collector.ReconcileServiceMonitor")
		return err
	}

	if err = factory.ReconcileCollectorConfig(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Reader, clusterRequest.Forwarder.Namespace, collectorConfig, clusterRequest.ResourceOwner); err != nil {
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

func (clusterRequest *ClusterLoggingRequest) addSecurityLabelsToNamespace() error {
	if clusterRequest.Forwarder.Namespace != constants.OpenshiftNS {
		return nil
	}
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
