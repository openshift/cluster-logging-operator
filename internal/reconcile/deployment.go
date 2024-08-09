package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/deployments"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment reconciles a Deployment to the desired spec returning an error
// if there is an issue creating or updating to the desired state
func Deployment(k8Client client.Client, desired *apps.Deployment) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &apps.Deployment{}
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v Deployment: %w", key, err)
		}
		same := false

		if same, _ = deployments.AreSame(current, desired); same {
			log.V(3).Info("Deployments are the same skipping update", "deploymentName", current.Name)
			return nil
		}
		current.Labels = desired.Labels
		current.Spec = desired.Spec
		current.OwnerReferences = desired.OwnerReferences
		return k8Client.Update(context.TODO(), current)
	})

	return retryErr
}
