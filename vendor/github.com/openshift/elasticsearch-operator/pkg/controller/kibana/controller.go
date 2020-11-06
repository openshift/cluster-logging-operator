package kibana

import (
	"time"

	loggingv1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/pkg/elasticsearch"
	"github.com/openshift/elasticsearch-operator/pkg/k8shandler"
	"github.com/openshift/elasticsearch-operator/pkg/k8shandler/kibana"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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
func (r *ReconcileKibana) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	es, err := k8shandler.GetElasticsearchCR(r.client, request.Namespace)
	if err != nil {
		logrus.Infof("skipping kibana reconciliation in %q: %s", request.Namespace, err)
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	esClient := elasticsearch.NewClient(es.Name, es.Namespace, r.client)
	if err := kibana.Reconcile(request, r.client, esClient); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logrus.Errorf("kibana reconcile err %v", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
