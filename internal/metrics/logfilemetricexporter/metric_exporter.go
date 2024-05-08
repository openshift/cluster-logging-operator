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
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func Reconcile(lfmeInstance *loggingv1alpha1.LogFileMetricExporter,
	requestClient client.Client,
	er record.EventRecorder,
	owner metav1.OwnerReference) error {

	// Adding common labels
	commonLabels := func(o runtime.Object) {
		runtime.SetCommonLabels(o, "lfme-service", constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName)
	}

	if err := reconcile.SecurityContextConstraints(requestClient, auth.NewSCC()); err != nil {
		log.V(9).Error(err, "logfilemetricexporter.SecurityContextConstraints")
		return err
	}

	resNames := &factory.ForwarderResourceNames{
		CommonName:                       constants.LogfilesmetricexporterName,
		ServiceAccount:                   constants.LogfilesmetricexporterName,
		ServiceAccountTokenSecret:        constants.LogfilesmetricexporterName + "-token",
		MetadataReaderClusterRoleBinding: "cluster-logging-" + constants.LogfilesmetricexporterName + "-metadata-reader",
	}

	if err := auth.ReconcileServiceAccount(er, requestClient, lfmeInstance.Namespace, resNames, owner); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileServiceAccount")
		return err
	}

	if err := auth.ReconcileRBAC(er, requestClient, constants.OpenshiftNS, resNames, owner); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileRBAC")
		return err
	}

	if err := network.ReconcileService(er, requestClient, lfmeInstance.Namespace, constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName, ExporterPortName, ExporterMetricsSecretName, ExporterPort, owner, commonLabels); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileService")
		return err
	}

	if err := metrics.ReconcileServiceMonitor(er, requestClient, lfmeInstance.Namespace, constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName, ExporterPortName, owner); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileServiceMonitor")
		return err
	}

	if err := ReconcileDaemonset(*lfmeInstance,
		er,
		requestClient,
		lfmeInstance.Namespace,
		constants.LogfilesmetricexporterName,
		owner,
		commonLabels); err != nil {
		msg := fmt.Sprintf("Unable to reconcile LogFileMetricExporter: %v", err)
		log.Error(err, msg)
		return errors.New(msg)
	}

	return nil
}
