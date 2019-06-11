package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewServiceAccount stubs a new instance of ServiceAccount
func NewServiceAccount(accountName string, namespace string) *core.ServiceAccount {
	return &core.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: core.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      accountName,
			Namespace: namespace,
		},
	}
}

//CreateOrUpdateServiceAccount creates or updates a ServiceAccount for logging with the given name
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateServiceAccount(name string) error {

	serviceAccount := NewServiceAccount(name, clusterRequest.cluster.Namespace)

	utils.AddOwnerRefToObject(serviceAccount, utils.AsOwner(clusterRequest.cluster))

	err := clusterRequest.Create(serviceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating %s service account: %v", name, err)
	}

	return nil
}

//RemoveServiceAccount of given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveServiceAccount(serviceAccountName string) error {

	serviceAccount := NewServiceAccount(serviceAccountName, clusterRequest.cluster.Namespace)

	err := clusterRequest.Delete(serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service account: %v", serviceAccountName, err)
	}

	return nil
}
