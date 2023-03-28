package logfilemetricsexporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ctrl "sigs.k8s.io/controller-runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ reconcile.Reconciler = &ReconcileLogFileMetricExporter{}

type ReconcileLogFileMetricExporter struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	//ClusterVersion is the semantic version of the cluster
	ClusterVersion string
	//ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string
}

var condReady = status.Condition{Type: loggingv1.ConditionReady, Status: corev1.ConditionTrue}

func condNotReady(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return loggingv1.NewCondition(loggingv1.ConditionReady, corev1.ConditionFalse, r, format, args...)
}

func (r *ReconcileLogFileMetricExporter) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("logfilemetricsexporter-controller fetching LFME instance")

	telemetry.SetLFMEMetrics(0) // Cancel previous info metric
	defer func() { telemetry.SetLFMEMetrics(1) }()

	lfmeInstance := loggingruntime.NewLogFileMetricExporter(request.NamespacedName.Namespace, request.NamespacedName.Name)
	r.Recorder.Event(lfmeInstance, corev1.EventTypeNormal, "ReconcilingLFMECR", "Reconciling Log File Metrics Exporter resource")

	if err := r.Client.Get(ctx, request.NamespacedName, lfmeInstance); err != nil {
		log.V(2).Info("logfilemetricsexporter-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err)
		if !errors.IsNotFound(err) {
			// Requeue the request
			return ctrl.Result{}, err
		}
		// Stop reconciliation
		return ctrl.Result{}, nil
	}

	// Check for singleton. Must be named instance
	if lfmeInstance.Name != constants.SingletonName {
		failMessage := fmt.Sprintf("Invalid name %q, singleton instance must be named %q",
			lfmeInstance.Name, constants.SingletonName)
		lfmeInstance.Status.Conditions.SetCondition(
			condNotReady(loggingv1.ReasonInvalid,
				failMessage))
		r.Recorder.Event(lfmeInstance, "Error", string(loggingv1.ReasonInvalid), failMessage)
		return r.updateStatus(lfmeInstance)
	}

	log.V(3).Info("logfilemetricexporter-controller run reconciler...")

	reconcileErr := k8shandler.ReconcileForLogFileMetricExporter(lfmeInstance, r.Client, r.Recorder, r.ClusterID, utils.AsOwner(lfmeInstance))

	if reconcileErr != nil {
		lfmeInstance.Status.Conditions.SetCondition(
			condNotReady(loggingv1.ReasonInvalid, reconcileErr.Error()))
		// if cluster is set to fail to reconcile then set healthStatus as 0
		telemetry.Data.LFMEInfo.Set(telemetry.HealthStatus, constants.UnHealthyStatus)
		log.V(2).Error(reconcileErr, "logfilemetricexporter-controller returning, error")

		r.Recorder.Event(lfmeInstance, "Error", string(loggingv1.ReasonInvalid), reconcileErr.Error())
	} else {
		if !lfmeInstance.Status.Conditions.SetCondition(condReady) {
			telemetry.Data.LFMEInfo.Set(telemetry.HealthStatus, constants.HealthyStatus)
			r.Recorder.Event(lfmeInstance, "Normal", string(condReady.Type), "LogFileMetricExporter deployed and ready")
		}
	}

	if result, err := r.updateStatus(lfmeInstance); err != nil {
		return result, err
	}

	return ctrl.Result{}, reconcileErr
}

func (r *ReconcileLogFileMetricExporter) updateStatus(instance *loggingv1alpha1.LogFileMetricExporter) (ctrl.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {

		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// do manual retry without error
			// more information about this error here: https://github.com/kubernetes/kubernetes/issues/28149
			return reconcile.Result{RequeueAfter: time.Second * 1}, nil
		}

		log.Error(err, "logfilemetricsexporter-controller error updating status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ReconcileLogFileMetricExporter) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		Watches(&source.Kind{Type: &loggingv1alpha1.LogFileMetricExporter{}}, &handler.EnqueueRequestForObject{})

	return controllerBuilder.
		For(&loggingv1alpha1.LogFileMetricExporter{}).
		Complete(r)
}
