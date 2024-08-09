package reconcile

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceAccount(k8Client client.Client, desired *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	current := &v1.ServiceAccount{}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		key := client.ObjectKeyFromObject(desired)
		if err := k8Client.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				current = desired
				return k8Client.Create(context.TODO(), desired)
			}
			return fmt.Errorf("failed to get %v ServiceAccount: %w", key, err)
		}

		same := false
		if same = (utils.AreMapsSame(current.Annotations, desired.Annotations) &&
			utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences)); same {
			log.V(3).Info("ServiceAccount are the same skipping update")
			return nil
		}
		if current.Annotations == nil {
			current.Annotations = map[string]string{}
		}
		if desired.Annotations != nil {
			for key, value := range desired.Annotations {
				current.Annotations[key] = value
			}
		}

		current.OwnerReferences = desired.OwnerReferences

		return k8Client.Update(context.TODO(), current)
	})
	return current, retryErr
}
