package proxyconfig

import (
	"time"

	"github.com/openshift/elasticsearch-operator/pkg/k8shandler/kibana"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	reconcilePeriod = 30 * time.Second
	reconcileResult = reconcile.Result{RequeueAfter: reconcilePeriod}
)

// Add creates a new ClusterLogging Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	if err := configv1.Install(mgr.GetScheme()); err != nil {
		return &ReconcileProxyConfig{}
	}

	return &ReconcileProxyConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("proxyconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for updates only
	pred := predicate.Funcs{
		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		CreateFunc:  func(e event.CreateEvent) bool { return true },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}

	// Watch for changes to the proxy resource.
	if err = c.Watch(&source.Kind{Type: &configv1.Proxy{}}, &handler.EnqueueRequestForObject{}, pred); err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileProxyConfig{}

// ReconcileProxyConfig reconciles a ClusterLogging object
type ReconcileProxyConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a cluster-scoped named "cluster" as well as
// trusted CA bundle configmap objects for the collector and the visualization resources.
// When the user configured and/or system certs are updated, the change is propagated to the
// configmap objects and this reconciler triggers to restart those pods.
func (r *ReconcileProxyConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if err := kibana.ReconcileKibanaInstance(request, r.client); err != nil {
		return reconcileResult, err
	}

	return reconcile.Result{}, nil
}
