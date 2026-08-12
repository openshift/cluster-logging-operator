package admission

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type saUsageAdmissionRunnable struct {
	client client.Client
}

// NewSAUsageAdmissionRunnable reconciles the SA usage ValidatingAdmissionPolicy once
// the manager (and its cache) has started.
func NewSAUsageAdmissionRunnable(k8sClient client.Client) *saUsageAdmissionRunnable {
	return &saUsageAdmissionRunnable{client: k8sClient}
}

func (r *saUsageAdmissionRunnable) Start(ctx context.Context) error {
	log.Info("Reconciling SA usage ValidatingAdmissionPolicy")
	if err := ReconcileSAUsageAuthorization(ctx, r.client); err != nil {
		log.Error(err, "unable to reconcile SA usage ValidatingAdmissionPolicy")
		return err
	}
	return nil
}

// NeedLeaderElection ensures only the elected operator pod manages cluster-scoped admission policy.
func (r *saUsageAdmissionRunnable) NeedLeaderElection() bool {
	return true
}
