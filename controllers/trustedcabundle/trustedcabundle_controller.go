package trustedcabundle

import (
	"context"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ reconcile.Reconciler = &ReconcileTrustedCABundle{}

//ReconcileTrustedCABundle reconciles the trusted CA bundle config map.
type ReconcileTrustedCABundle struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
}

// Reconcile reads that state of the trusted CA bundle configmap objects for the
// collector and the visualization resources.
// When the user configured and/or system certs are updated, the pods are triggered to restart.
func (r *ReconcileTrustedCABundle) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if err := k8shandler.ReconcileForTrustedCABundle(request.Name, r.client); err != nil {
		// Failed to reconcile - requeuing.
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// handleConfigMap returns true if meta namespace is "openshift-logging".
func handleConfigMap(meta metav1.Object) bool {
	return meta.GetNamespace() == constants.OpenshiftNS && utils.ContainsString(constants.ReconcileForGlobalProxyList, meta.GetName())
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileTrustedCABundle) SetupWithManager(mgr ctrl.Manager) error {
	// Watch for changes to the additional trust bundle configmap in "openshift-logging".
	pred := predicate.Funcs{
		UpdateFunc:  func(e event.UpdateEvent) bool { return handleConfigMap(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		CreateFunc:  func(e event.CreateEvent) bool { return handleConfigMap(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}

	controllerBuilder := ctrl.NewControllerManagedBy(mgr).Watches(&source.Kind{Type: &corev1.ConfigMap{}},
		&handler.EnqueueRequestForObject{},
		builder.WithPredicates(pred))

	return controllerBuilder.
		For(&loggingv1.ClusterLogging{}).
		Complete(r)
}
