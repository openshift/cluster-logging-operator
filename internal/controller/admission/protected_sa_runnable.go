package admission

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	internaladmission "github.com/openshift/cluster-logging-operator/internal/admission"
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
	err := wait.ExponentialBackoffWithContext(ctx, internaladmission.AdmissionReconcileBackoff, func(ctx context.Context) (bool, error) {
		lastErr = internaladmission.ReconcileProtectedSAPolicies(ctx, r.client, r.operatorNS)
		if lastErr == nil {
			return true, nil
		}
		if internaladmission.IsUnsupportedAdmissionPolicyAPI(lastErr) {
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

func (r *protectedSAAdmissionRunnable) NeedLeaderElection() bool {
	return true
}
