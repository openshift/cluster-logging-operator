package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"k8s.io/apimachinery/pkg/api/errors"
)

//RemoveService with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveService(serviceName string) error {

	service := runtime.NewService(clusterRequest.Cluster.Namespace, serviceName)
	err := clusterRequest.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}
