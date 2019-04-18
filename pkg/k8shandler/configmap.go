package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
