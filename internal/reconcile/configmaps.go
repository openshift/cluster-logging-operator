package reconcile

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/configmaps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Configmap creates or updates a ConfigMap owned by the desired ownerReferences.
// If a ConfigMap with the same name already exists and is not owned by the desired
// owners, it is left unchanged and an error is returned (LOG-9591).
func Configmap(k8Client client.Client, reader client.Reader, configMap *corev1.ConfigMap, opts ...comparators.ComparisonOption) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &corev1.ConfigMap{}
		key := client.ObjectKeyFromObject(configMap)
		if err := reader.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				return k8Client.Create(context.TODO(), configMap)
			}
			return fmt.Errorf("failed to get %v configmap: %v", key, err)
		}

		if err := utils.EnsureCanUpdateOwnedResource(current, configMap.OwnerReferences...); err != nil {
			return err
		}

		if configmaps.AreSame(current, configMap, opts...) {
			return nil
		}

		current.Data = configMap.Data
		current.Labels = configMap.Labels
		current.Annotations = configMap.Annotations
		current.OwnerReferences = configMap.OwnerReferences
		return k8Client.Update(context.TODO(), current)
	})
}
