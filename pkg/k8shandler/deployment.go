package k8shandler

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewDeployment stubs an instance of a Deployment
func NewDeployment(deploymentName string, namespace string, loggingComponent string, component string, podSpec core.PodSpec) *apps.Deployment {
	return &apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"provider":      "openshift",
				"component":     component,
				"logging-infra": loggingComponent,
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: utils.GetInt32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"provider":      "openshift",
					"component":     component,
					"logging-infra": loggingComponent,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: deploymentName,
					Labels: map[string]string{
						"provider":      "openshift",
						"component":     component,
						"logging-infra": loggingComponent,
					},
					Annotations: map[string]string{
						"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
					},
				},
				Spec: podSpec,
			},
			Strategy: apps.DeploymentStrategy{
				Type: apps.RollingUpdateDeploymentStrategyType,
				//RollingUpdate: {}
			},
		},
	}
}

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

	deployment := NewDeployment(
		deploymentName,
		clusterRequest.Cluster.Namespace,
		deploymentName,
		deploymentName,
		core.PodSpec{},
	)

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
