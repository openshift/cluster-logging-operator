package factory

import (
	"github.com/openshift/cluster-logging-operator/pkg/utils"

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
