package utils

import (
	"fmt"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
func RemoveConfigMap(namespace string, configmapName string) error {

	configMap := NewConfigMap(
		configmapName,
		namespace,
		map[string]string{},
	)

	err := sdk.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", configmapName, err)
	}

	return nil
}
