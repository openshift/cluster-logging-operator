package forwarding

import (
	"context"
	"time"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/go-logr/logr"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/telemetry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {

	logger := log.NewLogger("cluster-logging-operator")
	return &ReconcileForwarder{
		Log:      logger.WithName("clusterlogforwarder"),
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("clusterlogforwarder"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterlogforwarder-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterLogging
	err = c.Watch(&source.Kind{Type: &logging.ClusterLogForwarder{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileForwarder{}

// ReconcileForwarder reconciles a ClusterLogForwarder object
type ReconcileForwarder struct {
	// This Client, initialized using mgr.Client() above, is a split Client
	// that reads objects from the cache and writes to the apiserver
	Log      logr.Logger
	Client   client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var (
	reconcilePeriod = 30 * time.Second
	reconcileResult = reconcile.Result{RequeueAfter: reconcilePeriod}
)

var condReady = status.Condition{Type: logging.ConditionReady, Status: corev1.ConditionTrue}

func condNotReady(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionReady, corev1.ConditionFalse, r, format, args...)
}

func condDegraded(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionReady, corev1.ConditionTrue, r, format, args...)
}

func condInvalid(format string, args ...interface{}) status.Condition {
	return condNotReady(logging.ReasonInvalid, format, args...)
}

// Reconcile reads that state of the cluster for a ClusterLogForwarder object and makes changes based on the state read
// and what is in the Logging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileForwarder) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.Log.V(3).Info("clusterlogforwarder-controller fetching LF instance")

	// Fetch the ClusterLogForwarder instance
	instance := &logging.ClusterLogForwarder{}
	if err := r.Client.Get(ctx, request.NamespacedName, instance); err != nil {
		r.Log.V(2).Info("clusterlogforwarder-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err)
		if !errors.IsNotFound(err) {
			// Error reading - requeue the request.
			return reconcileResult, err
		}
		// else the object is not found -- meaning it was removed so stop reconciliation
		return reconcile.Result{}, nil
	}

	r.Log.V(3).Info("clusterlogforwarder-controller run reconciler...")

	reconcileErr := k8shandler.ReconcileForClusterLogForwarder(instance, r.Client)
	if reconcileErr != nil {
		// if cluster is set to fail to reconcile then set healthStatus as 0
		telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
		telemetry.UpdateCLFMetricsNoErr()
		r.Log.V(2).Error(reconcileErr, "clusterlogforwarder-controller returning, error")
	}

	//check for instancename and then update status
	if instance.Name != constants.SingletonName {
		instance.Status.Conditions.SetCondition(condInvalid("Invalid name %q, singleton instance must be named %q", instance.Name, constants.SingletonName))
		return r.updateStatus(instance)
	}

	if instance.Status.IsReady() {
		if instance.Status.Conditions.SetCondition(condReady) {
			telemetry.Data.CLFInfo.Set("healthStatus", constants.HealthyStatus)
			telemetry.UpdateCLFMetricsNoErr()
			r.Recorder.Event(instance, "Normal", string(condReady.Type), "All pipelines are valid")
		}
	}

	if instance.Status.IsDegraded() {
		msg := "Some pipelines are degraded or invalid"
		if instance.Status.Conditions.SetCondition(condDegraded(logging.ReasonInvalid, msg)) {
			telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
			telemetry.UpdateCLFMetricsNoErr()
			r.Recorder.Event(instance, "Error", string(logging.ReasonInvalid), msg)
		}
	}

	if result, err := r.updateStatus(instance); err != nil {
		return result, err
	}

	return reconcile.Result{}, reconcileErr
}

func (r *ReconcileForwarder) updateStatus(instance *logging.ClusterLogForwarder) (reconcile.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
		r.Log.Error(err, "clusterlogforwarder-controller error updating status")
		return reconcileResult, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileForwarder) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logging.ClusterLogForwarder{}).
		Complete(r)
}
