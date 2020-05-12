package kibana

import (
	"time"

	loggingv1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/pkg/k8shandler/kibana"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_kibana")

// Add creates a new Kibana Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKibana{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kibana-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Kibana
	err = c.Watch(&source.Kind{Type: &loggingv1.Kibana{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKibana implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKibana{}

// ReconcileKibana reconciles a Kibana object
type ReconcileKibana struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Kibana object and makes changes based on the state read
var (
	reconcilePeriod = 30 * time.Second
	reconcileResult = reconcile.Result{RequeueAfter: reconcilePeriod}
)

func (r *ReconcileKibana) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if err := kibana.ReconcileKibanaInstance(request, r.client); err != nil {
		return reconcileResult, err
	}

	return reconcile.Result{}, nil
}
