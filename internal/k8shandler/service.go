package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	core "k8s.io/api/core/v1"
)

// RemoveService with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveService(serviceName string) error {

	service := factory.NewService(
		serviceName,
		clusterRequest.Forwarder.Namespace,
		serviceName,
		[]core.ServicePort{},
	)

	err := clusterRequest.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}
