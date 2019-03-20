package utils

import (
	"fmt"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
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
func CreateOrUpdateSecret(secret *core.Secret) (err error) {
	err = sdk.Create(secret)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure constructing %v secret: %v", secret.Name, err)
		}

		current := secret.DeepCopy()
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = sdk.Get(current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v secret: %v", secret.Name, err)
			}

			current.Data = secret.Data
			if err = sdk.Update(current); err != nil {
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

//RemoveSecret with the given name in namespace
func RemoveSecret(namespace string, secretName string) error {

	secret := NewSecret(
		secretName,
		namespace,
		map[string][]byte{},
	)

	err := sdk.Delete(secret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v secret: %v", secretName, err)
	}

	return nil
}
