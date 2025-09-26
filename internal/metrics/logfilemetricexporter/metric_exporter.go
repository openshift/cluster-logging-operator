package logfilemetricexporter

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func Reconcile(lfmeInstance *loggingv1alpha1.LogFileMetricExporter,
	requestClient client.Client,
	uncachedReader client.Reader,
	owner metav1.OwnerReference) error {

	// Adding common labels
	commonLabels := func(o runtime.Object) {
		runtime.SetCommonLabels(o, constants.LogfilesmetricexporterName, lfmeInstance.Name, constants.LogfilesmetricexporterName)
	}

	if err := reconcile.SecurityContextConstraints(requestClient, uncachedReader, auth.NewSCC()); err != nil {
		log.V(9).Error(err, "logfilemetricexporter.SecurityContextConstraints")
		return err
	}

	resNames := &factory.ForwarderResourceNames{
		CommonName:                       constants.LogfilesmetricexporterName,
		ServiceAccount:                   constants.LogfilesmetricexporterName,
		ServiceAccountTokenSecret:        constants.LogfilesmetricexporterName + "-token",
		MetadataReaderClusterRoleBinding: "cluster-logging-" + constants.LogfilesmetricexporterName + "-metadata-reader",
	}

	if err := auth.ReconcileServiceAccount(requestClient, lfmeInstance.Namespace, resNames, owner); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileServiceAccount")
		return err
	}

	if err := auth.ReconcileRBAC(requestClient, resNames.CommonName, lfmeInstance.Namespace, resNames.ServiceAccount, owner); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileRBAC")
		return err
	}

	if err := network.ReconcileService(requestClient, lfmeInstance.Namespace, resNames.CommonName, lfmeInstance.Name, constants.LogfilesmetricexporterName, constants.MetricsPortName, ExporterMetricsSecretName, exporterPort, owner, commonLabels); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileService")
		return err
	}

	// Reconcile the network policy if it is defined
	if lfmeInstance.Spec.NetworkPolicy != nil {
		if err := network.ReconcileLogFileMetricsExporterNetworkPolicy(requestClient, lfmeInstance.Namespace, "lfme-"+resNames.CommonName, lfmeInstance.Name, constants.LogfilesmetricexporterName, lfmeInstance.Spec.NetworkPolicy.RuleSet, owner, commonLabels); err != nil {
			log.Error(err, "logfilemetricexporter.ReconcileNetworkPolicy")
			return err
		}
		// Attempt to remove the network policy if it is not defined
	} else {
		if err := network.RemoveNetworkPolicy(requestClient, lfmeInstance.Namespace, "lfme-"+resNames.CommonName); err != nil {
			log.Error(err, "logfilemetricexporter.RemoveNetworkPolicy")
			return err
		}
	}

	metricsSelector := metrics.BuildSelector(constants.LogfilesmetricexporterName, lfmeInstance.Name)
	if err := metrics.ReconcileServiceMonitor(requestClient, lfmeInstance.Namespace, resNames.CommonName, owner, metricsSelector, constants.MetricsPortName); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileServiceMonitor")
		return err
	}

	if err := ReconcileDaemonset(*lfmeInstance,
		requestClient,
		lfmeInstance.Namespace,
		resNames.CommonName,
		owner,
		commonLabels); err != nil {
		msg := fmt.Sprintf("Unable to reconcile LogFileMetricExporter: %v", err)
		log.Error(err, msg)
		return errors.New(msg)
	}

	return nil
}
