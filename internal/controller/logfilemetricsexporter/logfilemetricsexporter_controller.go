package logfilemetricsexporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	logmetricexporter "github.com/openshift/cluster-logging-operator/internal/metrics/logfilemetricexporter"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/validations/logfilemetricsexporter"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ReconcileLogFileMetricExporter{}

type ReconcileLogFileMetricExporter struct {
	Client client.Client
	Scheme *runtime.Scheme
	// ClusterVersion is the semantic version of the cluster
	ClusterVersion string
	// ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string
}

func condReady() metav1.Condition {
	return internalobs.NewCondition(observabilityv1.ConditionTypeReady, metav1.ConditionTrue, loggingv1alpha1.ReasonValid, "")
}

func condNotReady(r string, format string, args ...interface{}) metav1.Condition {
	return internalobs.NewCondition(observabilityv1.ConditionTypeReady, metav1.ConditionFalse, r, fmt.Sprintf(format, args...))
}

func (r *ReconcileLogFileMetricExporter) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("logfilemetricsexporter-controller fetching LFME instance")

	lfmeInstance := loggingruntime.NewLogFileMetricExporter(request.NamespacedName.Namespace, request.NamespacedName.Name)

	if err := r.Client.Get(ctx, request.NamespacedName, lfmeInstance); err != nil {
		log.V(2).Info("logfilemetricsexporter-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err)
		if !errors.IsNotFound(err) {
			// Requeue the request
			return ctrl.Result{}, err
		}
		// Stop reconciliation
		return ctrl.Result{}, nil
	}

	if lfmeInstance.DeletionTimestamp != nil {
		// Resource is being deleted, no further reconciliation
		return ctrl.Result{}, nil
	}

	// Validate LogFileMetricExporter instance
	if err, _ := logfilemetricsexporter.Validate(lfmeInstance); err != nil {
		condition := condNotReady(loggingv1alpha1.ReasonInvalid, "validation failed: %v", err)
		setCondition(&lfmeInstance.Status, condition)
		return r.updateStatus(lfmeInstance)
	}

	log.V(3).Info("logfilemetricexporter-controller run reconciler...")
	reconcileErr := logmetricexporter.Reconcile(lfmeInstance, r.Client, utils.AsOwner(lfmeInstance))

	if reconcileErr != nil {
		condition := condNotReady(loggingv1alpha1.ReasonInvalid, "%s", reconcileErr.Error())
		setCondition(&lfmeInstance.Status, condition)

		// if cluster is set to fail to reconcile then set healthStatus as 0
		log.V(2).Error(reconcileErr, "logfilemetricexporter-controller returning, error")
	} else {
		condition := condReady()
		setCondition(&lfmeInstance.Status, condition)
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

func setCondition(status *loggingv1alpha1.LogFileMetricExporterStatus, newCond metav1.Condition) bool {
	newCond.LastTransitionTime = metav1.Time{Time: time.Now()}

	for i, condition := range status.Conditions {
		if condition.Type == newCond.Type {
			if condition.Status == newCond.Status {
				newCond.LastTransitionTime = condition.LastTransitionTime
			}
			changed := condition.Status != newCond.Status ||
				condition.Reason != newCond.Reason ||
				condition.Message != newCond.Message
			status.Conditions[i] = newCond
			return changed
		}
	}

	status.Conditions = append(status.Conditions, newCond)
	return true
}
