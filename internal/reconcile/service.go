package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/services"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service reconciles a Service to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func Service(k8Client client.Client, desired *corev1.Service) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &corev1.Service{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v Service: %w", key, err)
		}
		same := false

		if same, _ = services.AreSame(current, desired); same {
			log.V(3).Info("Service is the same skipping update")
			return nil
		}

		//Explicitly copying because services are immutable
		current.Labels = desired.Labels
		current.Spec.Selector = desired.Spec.Selector
		current.Spec.Ports = desired.Spec.Ports
		current.OwnerReferences = desired.OwnerReferences
		return k8Client.Update(context.TODO(), current)
	})
	return retryErr
}
