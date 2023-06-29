package reconcile

import (
	"context"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/configmaps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Configmap(k8Client client.Writer, reader client.Reader, configMap *corev1.ConfigMap, opts ...configmaps.ComparisonOption) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &corev1.ConfigMap{}
		key := client.ObjectKeyFromObject(configMap)
		if err := reader.Get(context.TODO(), key, current); err != nil {
			if errors.IsNotFound(err) {
				return k8Client.Create(context.TODO(), configMap)
			}
			return fmt.Errorf("Failed to get %v configmap: %v", key, err)
		}
		if configmaps.AreSame(current, configMap, opts...) && utils.HasSameOwner(current.OwnerReferences, configMap.OwnerReferences) {
			return nil
		} else {
			current.Data = configMap.Data
			current.Labels = configMap.Labels
			current.Annotations = configMap.Annotations
			current.OwnerReferences = configMap.OwnerReferences
		}
		return k8Client.Update(context.TODO(), current)
	})
}
