package k8shandler

import (
	"fmt"
	"reflect"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

//NewConfigMap stubs an instance of Configmap
func NewConfigMap(configmapName string, namespace string, data map[string]string) *core.ConfigMap {
	return &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: namespace,
		},
		Data: data,
	}
}

func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateConfigMap(configMap *core.ConfigMap) error {
	err := clusterRequest.Create(configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing trusted CA bundle configmap: %v", err)
		}

		current := &core.ConfigMap{}

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
			return clusterRequest.Update(current)
		})
		return retryErr
	}
	return nil
}

//RemoveConfigMap with a given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveConfigMap(configmapName string) error {

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
