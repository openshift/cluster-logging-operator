package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NetworkPolicy creates or updates a NetworkPolicy returning an error
// if there is an issue creating or updating the NetworkPolicy to the desired state
func NetworkPolicy(k8Client client.Client, desired *networkingv1.NetworkPolicy) error {
	np := runtime.NewNetworkPolicy(desired.Namespace, desired.Name)
	op, err := controllerutil.CreateOrUpdate(context.TODO(), k8Client, np, func() error {
		np.Labels = desired.Labels
		np.Spec = desired.Spec
		np.OwnerReferences = desired.OwnerReferences
		return nil
	})

	if err == nil {
		log.V(3).Info(fmt.Sprintf("reconciled networkpolicy - operation: %s", op))
	}

	return err
}
