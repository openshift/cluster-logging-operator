package k8shandler

import (
	"fmt"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func (clusterRequest *ClusterLoggingRequest) GetSecret(secretName string) (*core.Secret, error) {
	secret := &core.Secret{}
	err := clusterRequest.Get(secretName, secret)
	return secret, err
}

//RemoveSecret with the given name in namespace
func (clusterRequest *ClusterLoggingRequest) RemoveSecret(secretName string) error {

	secret := NewSecret(
		secretName,
		clusterRequest.Cluster.Namespace,
		map[string][]byte{},
	)

	err := clusterRequest.Delete(secret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v secret: %v", secretName, err)
	}

	return nil
}
