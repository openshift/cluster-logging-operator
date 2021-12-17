package reconcile

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/configmaps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileConfigmap(k8Client client.Client, configMap *corev1.ConfigMap, opts ...configmaps.ComparisonOption) error {
	err := k8Client.Create(context.TODO(), configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating configmap: %v", err)
		}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			current := &corev1.ConfigMap{}
			key := client.ObjectKeyFromObject(configMap)
			if err = k8Client.Get(context.TODO(), key, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v configmap: %v", key, err)
			}

			if configmaps.AreSame(current, configMap, opts...) {
				return nil
			} else {
				current.Data = configMap.Data
				current.Labels = configMap.Labels
			}
			return k8Client.Update(context.TODO(), current)
		})
		return retryErr
	}
	return nil
}
