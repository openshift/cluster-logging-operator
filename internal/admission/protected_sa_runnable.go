package admission

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type protectedSAAdmissionRunnable struct {
	client     client.Client
	operatorNS string
}

// NewProtectedSAAdmissionRunnable reconciles the protected collector
// ServiceAccount ValidatingAdmissionPolicies (and their param ConfigMap) once
// the manager cache has started.
func NewProtectedSAAdmissionRunnable(k8sClient client.Client, operatorNS string) *protectedSAAdmissionRunnable {
	return &protectedSAAdmissionRunnable{client: k8sClient, operatorNS: operatorNS}
}

func (r *protectedSAAdmissionRunnable) Start(ctx context.Context) error {
	log.Info("Reconciling protected collector ServiceAccount ValidatingAdmissionPolicies")

	var lastErr error
	err := wait.ExponentialBackoffWithContext(ctx, admissionReconcileBackoff, func(ctx context.Context) (bool, error) {
		lastErr = ReconcileProtectedSAPolicies(ctx, r.client, r.operatorNS)
		if lastErr == nil {
			return true, nil
		}
		if isUnsupportedAdmissionPolicyAPI(lastErr) {
			lastErr = nil
			return true, nil
		}
		log.V(1).Info("retrying protected SA ValidatingAdmissionPolicy reconciliation", "error", lastErr)
		return false, nil
	})
	if err != nil && !wait.Interrupted(err) {
		if lastErr != nil {
			log.Error(lastErr, "unable to reconcile protected SA ValidatingAdmissionPolicies", "reason", err)
		}
		return nil
	}
	if lastErr != nil {
		log.Error(lastErr, "unable to reconcile protected SA ValidatingAdmissionPolicies after retries")
	}
	return nil
}

// NeedLeaderElection ensures only the elected operator pod manages the
// cluster-scoped admission policies.
func (r *protectedSAAdmissionRunnable) NeedLeaderElection() bool {
	return true
}
