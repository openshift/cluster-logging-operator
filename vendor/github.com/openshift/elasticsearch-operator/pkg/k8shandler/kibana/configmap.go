package kibana

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

//NewConfigMap stubs an instance of Configmap
func NewConfigMap(configmapName string, namespace string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: data,
	}
}

func (clusterRequest *KibanaRequest) CreateOrUpdateConfigMap(configMap *corev1.ConfigMap) error {
	err := clusterRequest.Create(configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating configmap: %v", err)
		}

		current := &corev1.ConfigMap{}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(configMap.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v configmap for %q: %v", configMap.Name, clusterRequest.cluster.Name, err)
			}

			if reflect.DeepEqual(configMap.Data, current.Data) {
				return nil
			}
			current.Data = configMap.Data

			changed := false
			// if configMap specified labels ensure that current has them...
			if len(configMap.ObjectMeta.Labels) > 0 {
				for key, val := range configMap.ObjectMeta.Labels {
					if currentVal, ok := current.ObjectMeta.Labels[key]; ok {
						if currentVal != val {
							current.ObjectMeta.Labels[key] = val
							changed = true
						}
					} else {
						current.ObjectMeta.Labels[key] = val
						changed = true
					}
				}
			} else {
				return nil
			}
			if !changed {
				// shortcut updating -- we didn't change anything
				return nil
			}

			return clusterRequest.Update(current)
		})
		return retryErr
	}
	return nil
}

//RemoveConfigMap with a given name and namespace
func (clusterRequest *KibanaRequest) RemoveConfigMap(configmapName string) error {

	configMap := NewConfigMap(
		configmapName,
		clusterRequest.cluster.Namespace,
		map[string]string{},
	)

	err := clusterRequest.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", configmapName, err)
	}

	return nil
}
