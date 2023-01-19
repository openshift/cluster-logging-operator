package k8shandler

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConfigMap stubs an instance of Configmap
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

// RemoveConfigMap with a given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveConfigMap(configmapName string) error {

	configMap := NewConfigMap(
		configmapName,
		clusterRequest.Cluster.Namespace,
		map[string]string{},
	)

	err := clusterRequest.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", configmapName, err)
	}

	return nil
}
