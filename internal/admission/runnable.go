package admission

import (
	"context"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var saUsageReconcileBackoff = wait.Backoff{
	Steps:    5,
	Duration: 2 * time.Second,
	Factor:   2.0,
	Cap:      30 * time.Second,
}

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

	var lastErr error
	err := wait.ExponentialBackoffWithContext(ctx, saUsageReconcileBackoff, func(ctx context.Context) (bool, error) {
		lastErr = ReconcileSAUsageAuthorization(ctx, r.client)
		if lastErr == nil {
			return true, nil
		}
		if isUnsupportedAdmissionPolicyAPI(lastErr) {
			lastErr = nil
			return true, nil
		}
		log.V(1).Info("retrying SA usage ValidatingAdmissionPolicy reconciliation", "error", lastErr)
		return false, nil
	})
	if err != nil && !wait.Interrupted(err) {
		if lastErr != nil {
			log.Error(lastErr, "unable to reconcile SA usage ValidatingAdmissionPolicy", "reason", err)
		}
		return nil
	}
	if lastErr != nil {
		log.Error(lastErr, "unable to reconcile SA usage ValidatingAdmissionPolicy after retries")
	}
	return nil
}

// NeedLeaderElection ensures only the elected operator pod manages cluster-scoped admission policy.
func (r *saUsageAdmissionRunnable) NeedLeaderElection() bool {
	return true
}
