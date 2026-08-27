package admission

import (
	"context"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	internaladmission "github.com/openshift/cluster-logging-operator/internal/admission"
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

	backoff := internaladmission.AdmissionReconcileBackoff
	for {
		err := internaladmission.ReconcileProtectedSAPolicies(ctx, r.client, r.operatorNS)
		if err == nil {
			return nil
		}
		if internaladmission.IsUnsupportedAdmissionPolicyAPI(err) {
			return nil
		}
		delay := backoff.Step()
		log.V(1).Info("retrying protected SA ValidatingAdmissionPolicy reconciliation", "error", err, "retryAfter", delay)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(delay):
		}
	}
}

func (r *protectedSAAdmissionRunnable) NeedLeaderElection() bool {
	return true
}
