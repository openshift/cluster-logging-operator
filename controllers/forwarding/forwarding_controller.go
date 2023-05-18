package forwarding

import (
	"context"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"

	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	validationerrors "github.com/openshift/cluster-logging-operator/internal/validations/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ reconcile.Reconciler = &ReconcileForwarder{}

// ReconcileForwarder reconciles a ClusterLogForwarder object
type ReconcileForwarder struct {
	// This Client, initialized using mgr.Client() above, is a split Client
	// that reads objects from the cache and writes to the apiserver
	Client   client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	//ClusterVersion is the semantic version of the cluster
	ClusterVersion string
	//ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string
}

// Reconcile reads that state of the cluster for a ClusterLogForwarder object and makes changes based on the state read
// and what is in the Logging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileForwarder) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("clusterlogforwarder-controller fetching LF instance")
	r.Recorder.Event(loggingruntime.NewClusterLogForwarder(request.NamespacedName.Namespace, request.NamespacedName.Name), corev1.EventTypeNormal, constants.EventReasonReconcilingLoggingCR, "Reconciling logging resource")

	telemetry.SetCLFMetrics(0) // Cancel previous info metric
	defer func() { telemetry.SetCLFMetrics(1) }()

	// CL only needed if Forwarder is named instance
	var cl logging.ClusterLogging
	if request.Name == constants.SingletonName {
		var err error
		cl, err = loader.FetchClusterLogging(r.Client, constants.WatchNamespace, constants.SingletonName, false)
		if err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	resourceNames := factory.GenerateResourceNames(request.NamespacedName.Name, request.NamespacedName.Namespace)

	// Fetch the ClusterLogForwarder instance
	instance, err, status := loader.FetchClusterLogForwarder(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name, resourceNames.InternalLogStoreSecret, true, func() logging.ClusterLogging { return cl })
	if status != nil {
		instance.Status = *status
	}
	if err != nil {
		log.V(3).Info("clusterlogforwarder-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err.Error())
		if validationerrors.IsValidationError(err) {
			condition := logging.CondInvalid("validation failed: %v", err)
			instance.Status.Conditions.SetCondition(condition)
			r.Recorder.Event(&instance, "Warning", string(logging.ReasonInvalid), condition.Message)
			telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
			return r.updateStatus(&instance)
		} else if !errors.IsNotFound(err) {
			// Error reading - requeue the request.
			return ctrl.Result{}, err
		}
		// else the object is not found -- meaning it was removed so stop reconciliation
		return ctrl.Result{}, nil
	}

	log.V(3).Info("clusterlogforwarder-controller run reconciler...")

	reconcileErr := k8shandler.ReconcileForClusterLogForwarder(&instance, &cl, r.Client, r.Recorder, r.ClusterID, resourceNames)
	if reconcileErr != nil {
		// if cluster is set to fail to reconcile then set healthStatus as 0
		telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.V(2).Error(reconcileErr, "clusterlogforwarder-controller returning, error")
		logging.SetCondition(&instance.Status.Conditions, logging.CollectorDeadEnd, corev1.ConditionTrue, logging.ReasonInvalid, "error reconciling clusterlogforwarder instance: %v", reconcileErr)
	} else {
		if instance.Status.Conditions.IsTrueFor(logging.ConditionReady) {
			logging.SetCondition(&instance.Status.Conditions, logging.CollectorDeadEnd, corev1.ConditionFalse, "", "")
			// This returns False if SetCondition updates the condition instead of setting it.
			// For condReady, it will always be updating the status.
			if !instance.Status.Conditions.SetCondition(logging.CondReady) {
				telemetry.Data.CLFInfo.Set("healthStatus", constants.HealthyStatus)
				r.Recorder.Event(&instance, "Normal", string(logging.CondReady.Type), "ClusterLogForwarder is valid")
			}
		}
	}

	if result, err := r.updateStatus(&instance); err != nil {
		return result, err
	}

	return ctrl.Result{}, reconcileErr
}

func (r *ReconcileForwarder) updateStatus(instance *logging.ClusterLogForwarder) (ctrl.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {

		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// do manual retry without error
			// more information about this error here: https://github.com/kubernetes/kubernetes/issues/28149
			return reconcile.Result{RequeueAfter: time.Second * 1}, nil
		}

		log.Error(err, "clusterlogforwarder-controller error updating status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileForwarder) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		Watches(&source.Kind{Type: &logging.ClusterLogForwarder{}}, &handler.EnqueueRequestForObject{})
	return controllerBuilder.
		For(&logging.ClusterLogForwarder{}).
		Complete(r)
}
