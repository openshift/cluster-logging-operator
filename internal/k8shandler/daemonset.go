package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/factory"

	"k8s.io/apimachinery/pkg/api/errors"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDaemonSet stubs an instance of a daemonset
func NewDaemonSet(daemonsetName, namespace, loggingComponent, component, impl string, podSpec core.PodSpec) *apps.DaemonSet {
	return factory.NewDaemonSet(daemonsetName, namespace, loggingComponent, component, impl, podSpec)
}

// GetDaemonSetList lists DS in namespace with given selector
func (clusterRequest *ClusterLoggingRequest) GetDaemonSetList(selector map[string]string) (*apps.DaemonSetList, error) {
	list := &apps.DaemonSetList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(
		selector,
		list,
	)

	return list, err
}

// RemoveDaemonset with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveDaemonset(daemonsetName string) error {

	daemonset := NewDaemonSet(
		daemonsetName,
		clusterRequest.Cluster.Namespace,
		daemonsetName,
		daemonsetName,
		"vector", //impl does not matter here
		core.PodSpec{},
	)

	err := clusterRequest.Delete(daemonset)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v daemonset %v", daemonsetName, err)
	}

	return nil
}
