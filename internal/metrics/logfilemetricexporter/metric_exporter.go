package logfilemetricexporter

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	"github.com/openshift/cluster-logging-operator/internal/network"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

func Reconcile(lfmeInstance *loggingv1alpha1.LogFileMetricExporter,
	requestClient client.Client,
	er record.EventRecorder,
	clInstance loggingv1.ClusterLogging,
	owner metav1.OwnerReference) error {

	// Adding common labels
	commonLabels := func(o runtime.Object) {
		runtime.SetCommonLabels(o, "lfme-service", lfmeInstance.Name, constants.LogfilesmetricexporterName)
	}

	if err := network.ReconcileService(er, requestClient, lfmeInstance.Namespace, constants.LogfilesmetricexporterName, ExporterPortName, ExporterMetricsSecretName, ExporterPort, owner, commonLabels); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileService")
		return err
	}

	if err := metrics.ReconcileServiceMonitor(er, requestClient, lfmeInstance.Namespace, constants.LogfilesmetricexporterName, ExporterPortName, owner); err != nil {
		log.Error(err, "logfilemetricexporter.ReconcileServiceMonitor")
		return err
	}

	if err := ReconcileDaemonset(*lfmeInstance,
		er,
		requestClient,
		lfmeInstance.Namespace,
		constants.LogfilesmetricexporterName,
		clInstance.Spec.Collection.Type,
		owner,
		commonLabels); err != nil {
		msg := fmt.Sprintf("Unable to reconcile LogFileMetricExporter: %v", err)
		telemetry.Data.LFMEInfo.Set(telemetry.HealthStatus, constants.UnHealthyStatus)
		log.Error(err, msg)
		return errors.New(msg)
	}

	telemetry.Data.LFMEInfo.Set(telemetry.Deployed, telemetry.IsPresent)

	return nil
}
