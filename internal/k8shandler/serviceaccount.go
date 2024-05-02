package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RemoveServiceAccount of given name and namespace
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
