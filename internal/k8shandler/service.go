package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	core "k8s.io/api/core/v1"
)

// RemoveService with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveService(serviceName string) error {

	service := factory.NewService(
		serviceName,
		clusterRequest.Forwarder.Namespace,
		serviceName,
		constants.CollectorName,
		[]core.ServicePort{},
	)

	err := clusterRequest.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}

// RemoveInputService with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveInputService(serviceName string) error {
	service := runtime.NewService(clusterRequest.Forwarder.Namespace, serviceName)
	err := clusterRequest.Delete(service)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v service %v", serviceName, err)
	}

	return nil
}
