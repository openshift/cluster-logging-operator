package configmap

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/openshift/cluster-logging-operator/pkg/client/k8s"
)

//New stubs an instance of Configmap
func New(configmapName string, namespace string, data map[string]string) *corev1.ConfigMap {
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

//Reconcile creates a new config map resource unless it exists whereas it will update
//the existing config map if the data section changed.
func Reconcile(client k8s.Client, configMap *corev1.ConfigMap) error {
	err := client.Create(configMap)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating configmap: %v", err)
		}

		current := &corev1.ConfigMap{}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = client.Get(configMap.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v configmap: %v", configMap.Name, err)
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

			return client.Update(current)
		})
		return retryErr
	}
	return nil
}

//Remove configmap with a given name and namespace
func Remove(client k8s.Client, namespace, name string, HasCLORef func(object metav1.Object) bool) error {

	configMap := New(
		name,
		namespace,
		map[string]string{},
	)

	//TODO: Remove this in the next release after removing old kibana code completely
	if !HasCLORef(configMap) {
		return nil
	}

	err := client.Delete(configMap)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v configmap: %v", name, err)
	}

	return nil
}
