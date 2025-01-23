package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Service reconciles a Service to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func Service(k8Client client.Client, desired *corev1.Service) error {
	sm := runtime.NewService(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, sm, func() error {

		// Set annotations upon creation
		if sm.CreationTimestamp.IsZero() {
			sm.Annotations = desired.Annotations
		}

		sm.Labels = desired.Labels
		sm.Spec.Selector = desired.Spec.Selector
		sm.Spec.Ports = desired.Spec.Ports
		sm.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled service - operation: %s", op))
	}
	return err
}
