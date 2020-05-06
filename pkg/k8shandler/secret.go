package k8shandler

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewSecret stubs an instance of a secret
func NewSecret(secretName string, namespace string, data map[string][]byte) *core.Secret {
	return &core.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: "Opaque",
		Data: data,
	}
}

//CreateOrUpdateSecret creates or updates a secret and retries on conflict
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateSecret(secret *core.Secret) (err error) {
	err = clusterRequest.Create(secret)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v secret: %v", secret.Name, err)
		}

		current := &core.Secret{}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(secret.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v secret: %v", secret.Name, err)
			}
			if reflect.DeepEqual(current.Data, secret.Data) {
				// identical; no need to update.
				return nil
			}
			current.Data = secret.Data
			if err = clusterRequest.Update(current); err != nil {
				return err
			}
			return nil
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) GetSecret(secretName string) (*core.Secret, error) {
	secret := &core.Secret{}
	err := clusterRequest.Get(secretName, secret)
	return secret, err
}

//RemoveSecret with the given name in namespace
func (clusterRequest *ClusterLoggingRequest) RemoveSecret(secretName string) error {

	secret := NewSecret(
		secretName,
		clusterRequest.cluster.Namespace,
		map[string][]byte{},
	)

	err := clusterRequest.Delete(secret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v secret: %v", secretName, err)
	}

	return nil
}
