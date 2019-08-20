package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

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

	if serviceAccountName == "logcollector" {
		// remove our finalizer from the list and update it.
		serviceAccount.ObjectMeta.Finalizers = utils.RemoveString(serviceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
	}

	err := clusterRequest.Delete(serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service account: %v", serviceAccountName, err)
	}

	return nil
}

func NewLogCollectorServiceAccountRef(uid types.UID) metav1.OwnerReference {
    return metav1.OwnerReference {
        APIVersion:         "v1", // apiversion for serviceaccounts/finalizers in cluster-logging.<VER>.clusterserviceversion.yaml
        Kind:               "ServiceAccount",
        Name:               "logcollector",
        UID:                uid,
        BlockOwnerDeletion: utils.GetBool(true),
        Controller:         utils.GetBool(true),
    }
}
