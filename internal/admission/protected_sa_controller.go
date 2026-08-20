package admission

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProtectedSAReconciler keeps the protected-ServiceAccount param ConfigMap in
// sync with the set of ClusterLogForwarders. It reconciles on every CLF event
// (create/update/delete) by rebuilding the ConfigMap from the full CLF list, so
// deletions are handled without finalizers and the ConfigMap self-heals.
type ProtectedSAReconciler struct {
	Client     client.Client
	OperatorNS string
}

func (r *ProtectedSAReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if err := SyncProtectedServiceAccounts(ctx, r.Client, r.OperatorNS); err != nil {
		log.V(1).Error(err, "failed to sync protected collector ServiceAccounts", "trigger", req.NamespacedName)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ProtectedSAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&obsv1.ClusterLogForwarder{}).
		Named("protected-sa-configmap").
		Complete(r)
}
