package trustedcabundle

import (
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		return &ReconcileTrustedCABundle{}
	}

	return &ReconcileTrustedCABundle{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("trustedcabundle-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to the additional trust bundle configmap in "openshift-logging".
	pred := predicate.Funcs{
		UpdateFunc:  func(e event.UpdateEvent) bool { return handleConfigMap(e.MetaNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		CreateFunc:  func(e event.CreateEvent) bool { return handleConfigMap(e.Meta) },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
	if err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{}, pred); err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTrustedCABundle{}

// ReconcileProxyConfig reconciles a ClusterLogging object
type ReconcileTrustedCABundle struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the trusted CA bundle configmap objects for the
// collector and the visualization resources.
// When the user configured and/or system certs are updated, the pods are triggered to restart.
func (r *ReconcileTrustedCABundle) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// do one for fluentd and one for kibana separate...
	if utils.ContainsString(constants.ReconcileForGlobalProxyList, request.Name) {

		if err := k8shandler.ReconcileForTrustedCABundle(request.Name, r.client); err != nil {
			// Failed to reconcile - requeuing.
			return reconcileResult, err
		}
	}

	return reconcile.Result{}, nil
}

// handleConfigMap returns true if meta namespace is "openshift-logging".
func handleConfigMap(meta metav1.Object) bool {
	return meta.GetNamespace() == constants.OpenshiftNS && utils.ContainsString(constants.ReconcileForGlobalProxyList, meta.GetName())
}
