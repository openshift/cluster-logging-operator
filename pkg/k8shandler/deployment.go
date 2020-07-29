package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/factory"
	"k8s.io/apimachinery/pkg/api/errors"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetDeploymentList returns a list for a give namespace and selector
func (clusterRequest *ClusterLoggingRequest) GetDeploymentList(selector map[string]string) (*apps.DeploymentList, error) {
	list := &apps.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(
		selector,
		list,
	)

	return list, err
}

//RemoveDeployment of given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveDeployment(deploymentName string) error {

	deployment := factory.NewDeployment(
		deploymentName,
		clusterRequest.cluster.Namespace,
		deploymentName,
		deploymentName,
		core.PodSpec{},
	)

	//TODO: Remove this in the next release after removing old kibana code completely
	if !HasCLORef(deployment, clusterRequest) {
		return nil
	}

	err := clusterRequest.Delete(deployment)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v deployment %v", deploymentName, err)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) GetReplicaSetList(selector map[string]string) (*apps.ReplicaSetList, error) {
	list := &apps.ReplicaSetList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(
		selector,
		list,
	)

	return list, err
}
