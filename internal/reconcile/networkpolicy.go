package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NetworkPolicy creates a NetworkPolicy returning an error
// if there is an issue creating the NetworkPolicy
func NetworkPolicy(k8Client client.Client, desired *networkingv1.NetworkPolicy) error {
	np := runtime.NewNetworkPolicy(desired.Namespace, desired.Name)
	nsName := types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}
	err := k8Client.Get(context.TODO(), nsName, np)

	if err != nil {
		if errors.IsNotFound(err) {
			np.Labels = desired.Labels
			np.Spec = desired.Spec
			np.OwnerReferences = desired.OwnerReferences
			log.V(3).Info(fmt.Sprintf("created networkpolicy: %s", err))

			return k8Client.Create(context.TODO(), np)
		}
		return err
	}

	return err
}
