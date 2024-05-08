package logfilemetricsexporter

import (
	"context"
	"github.com/openshift/cluster-logging-operator/internal/validations/logfilemetricsexporter"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
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

	// Validate LogFileMetricExporter instance
	if err, _ := logfilemetricsexporter.Validate(lfmeInstance); err != nil {
		condition := loggingv1.CondInvalid("validation failed: %v", err)
		lfmeInstance.Status.Conditions.SetCondition(condition)
		r.Recorder.Event(lfmeInstance, corev1.EventTypeWarning, string(loggingv1.ReasonInvalid), condition.Message)
		return r.updateStatus(lfmeInstance)
	}

	log.V(3).Info("logfilemetricexporter-controller run reconciler...")
	reconcileErr := k8shandler.ReconcileForLogFileMetricExporter(lfmeInstance, r.Client, r.Recorder, r.ClusterID, utils.AsOwner(lfmeInstance))

	if reconcileErr != nil {
		lfmeInstance.Status.Conditions.SetCondition(
			condNotReady(loggingv1.ReasonInvalid, reconcileErr.Error()))
		log.V(2).Error(reconcileErr, "logfilemetricexporter-controller returning, error")

		r.Recorder.Event(lfmeInstance, corev1.EventTypeWarning, string(loggingv1.ReasonInvalid), reconcileErr.Error())
	} else {
		if !lfmeInstance.Status.Conditions.SetCondition(condReady) {
			r.Recorder.Event(lfmeInstance, corev1.EventTypeNormal, string(condReady.Type), "LogFileMetricExporter deployed and ready")
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1alpha1.LogFileMetricExporter{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.Service{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}
