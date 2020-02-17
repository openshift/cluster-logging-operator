package k8shandler

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
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
	if err := clusterRequest.Get(secretName, secret); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		return nil, fmt.Errorf("Failed to get %v secret: %v", secret.Name, err)
	}

	return secret, nil
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

func calcSecretHashValue(secret *core.Secret) (string, error) {
	hashValue := ""
	var err error

	if secret == nil {
		return hashValue, nil
	}

	hashKeys := []string{}
	rawbytes := []byte{}

	// we just want the keys here to sort them for consistently calculated hashes
	for key := range secret.Data {
		hashKeys = append(hashKeys, key)
	}

	sort.Strings(hashKeys)

	for _, key := range hashKeys {
		rawbytes = append(rawbytes, secret.Data[key]...)
	}

	hashValue, err = utils.CalculateMD5Hash(string(rawbytes))
	if err != nil {
		return "", err
	}

	return hashValue, nil
}
