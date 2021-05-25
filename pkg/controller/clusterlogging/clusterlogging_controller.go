package clusterlogging

import (
	"context"
	"time"

	"github.com/ViaQ/logerr/log"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new ClusterLogging Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterLogging{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor("clusterlogging-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterlogging-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterLogging
	err = c.Watch(&source.Kind{Type: &loggingv1.ClusterLogging{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterLogging{}

// ReconcileClusterLogging reconciles a ClusterLogging object
type ReconcileClusterLogging struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

var (
	reconcilePeriod = 30 * time.Second
	reconcileResult = reconcile.Result{RequeueAfter: reconcilePeriod}
)

// Reconcile reads that state of the cluster for a ClusterLogging object and makes changes based on the state read
// and what is in the ClusterLogging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterLogging) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.V(3).Info("Clusterlogging reconcile request.", "name", request.Name)

	// Fetch the ClusterLogging instance
	instance := &loggingv1.ClusterLogging{}
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

	if instance.Spec.ManagementState == loggingv1.ManagementStateUnmanaged {
		return reconcile.Result{}, nil
	}

	if err = k8shandler.Reconcile(instance, r.client, r.recorder); err != nil {
		log.Error(err, "Error reconciling clusterlogging instance")
	}

	if result, err := r.updateStatus(instance); err != nil {
		return result, err
	}

	return reconcileResult, err
}

func (r *ReconcileClusterLogging) updateStatus(instance *loggingv1.ClusterLogging) (reconcile.Result, error) {
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		log.Error(err, "clusterlogging-controller error updating status")
		return reconcileResult, err
	}

	return reconcile.Result{}, nil
}
