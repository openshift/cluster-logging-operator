package logforwarding

import (
	"context"
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_logforwarding")

// Add creates a new LogForwarding Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLogForwarding{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("logforwarding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for creates only
	pred := predicate.Funcs{
		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		CreateFunc:  func(e event.CreateEvent) bool { return true },
		GenericFunc: func(e event.GenericEvent) bool { return true },
	}

	// Watch for changes to primary resource LogForwarding
	err = c.Watch(&source.Kind{Type: &loggingv1alpha1.LogForwarding{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLogForwarding implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLogForwarding{}

// ReconcileLogForwarding reconciles a LogForwarding object
type ReconcileLogForwarding struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a LogForwarding object and makes changes based on the state read
// and what is in the LogForwarding.Spec
//
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLogForwarding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling LogForwarding")

	// Fetch the LogForwarding instance
	instance := &loggingv1alpha1.LogForwarding{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	clf := &loggingv1.ClusterLogForwarder{}
	if err := instance.ConvertTo(clf); err != nil {
		if err := r.updateStatus(request, fmt.Sprintf("%s", err)); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, err
	}

	createOpts := []client.CreateOption{}
	err = r.client.Create(context.TODO(), clf, createOpts...)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// ClusterLogForwarder already exists, don't requeue
			reqLogger.Info("Skip reconcile: ClusterLogFowarder already exists")
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	msg := fmt.Sprintf(
		"LogForwarding resource %q is replaced by ClusterLogForwarder resource `instance`",
		instance.GetName(),
	)
	r.updateStatus(request, msg)

	reqLogger.Info("Successfully converted LogForwarding to ClusterLogForwarder CR")
	return reconcile.Result{}, nil
}

func (r *ReconcileLogForwarding) updateStatus(request reconcile.Request, msg string) error {
	instance := &loggingv1alpha1.LogForwarding{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Mark the old LogForwarding CR as obsolete before converting to a ClusterLogForwarder CR
	instance.Status = &loggingv1alpha1.ForwardingStatus{
		State:   loggingv1alpha1.LogForwardingStateObsolete,
		Reason:  loggingv1alpha1.LogForwardingReasonObsolete,
		Message: msg,
	}

	updateOpts := []client.UpdateOption{}
	err = r.client.Status().Update(context.TODO(), instance, updateOpts...)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}
