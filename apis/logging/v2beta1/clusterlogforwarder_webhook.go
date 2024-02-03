package v2beta1

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ClusterLogForwarder) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
