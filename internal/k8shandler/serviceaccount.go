package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreateOrUpdateServiceAccount creates or updates a ServiceAccount for logging with the given name
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateServiceAccount(name string, annotations *map[string]string) error {

	serviceAccount := runtime.NewServiceAccount(clusterRequest.Cluster.Namespace, name)
	if annotations != nil {
		if serviceAccount.GetObjectMeta().GetAnnotations() == nil {
			serviceAccount.GetObjectMeta().SetAnnotations(make(map[string]string))
		}
		for key, value := range *annotations {
			serviceAccount.GetObjectMeta().GetAnnotations()[key] = value
		}
	}

	utils.AddOwnerRefToObject(serviceAccount, utils.AsOwner(clusterRequest.Cluster))

	if err := clusterRequest.Create(serviceAccount); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating %v serviceaccount: %v", serviceAccount.Name, err)
		}

		current := &core.ServiceAccount{}
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err = clusterRequest.Get(serviceAccount.Name, current); err != nil {
				if errors.IsNotFound(err) {
					// the object doesn't exist -- it was likely culled
					// recreate it on the next time through if necessary
					return nil
				}
				return fmt.Errorf("Failed to get %v serviceaccount: %v", serviceAccount.Name, err)
			}
			if annotations != nil && serviceAccount.GetObjectMeta().GetAnnotations() != nil {
				if current.GetObjectMeta().GetAnnotations() == nil {
					current.GetObjectMeta().SetAnnotations(make(map[string]string))
				}
				for key, value := range serviceAccount.GetObjectMeta().GetAnnotations() {
					current.GetObjectMeta().GetAnnotations()[key] = value
				}
			}
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

//RemoveServiceAccount of given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveServiceAccount(serviceAccountName string) error {

	serviceAccount := runtime.NewServiceAccount(clusterRequest.Cluster.Namespace, serviceAccountName)

	if serviceAccountName == constants.CollectorServiceAccountName {
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
	return metav1.OwnerReference{
		APIVersion:         "v1", // apiversion for serviceaccounts/finalizers in cluster-logging.<VER>.clusterserviceversion.yaml
		Kind:               "ServiceAccount",
		Name:               constants.CollectorServiceAccountName,
		UID:                uid,
		BlockOwnerDeletion: utils.GetBool(true),
		Controller:         utils.GetBool(true),
	}
}
