package k8shandler

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	client "sigs.k8s.io/controller-runtime/pkg/client"

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

// GetServiceList returns a list of services based on a key/value label and namespace
func (clusterRequest *ClusterLoggingRequest) GetServiceList(key, val, namespace string) (*core.ServiceList, error) {
	labelSelector, _ := labels.Parse(fmt.Sprintf("%s=%s", key, val))
	httpServices := core.ServiceList{}
	if err := clusterRequest.Reader.List(context.TODO(), &httpServices, &client.ListOptions{LabelSelector: labelSelector, Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("failure listing services with label: %s,  %v", fmt.Sprintf("%s=%s", key, val), err)
	}
	return &httpServices, nil
}
