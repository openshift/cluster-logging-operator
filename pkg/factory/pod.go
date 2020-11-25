package factory

import (
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	core "k8s.io/api/core/v1"
)

// NewPodSpec is a constructor to instaniate a new PodSpec.
// Notice that all Aggregated Logging relevant pods are (force-)allocated to linux nodes, see https://jira.coreos.com/browse/LOG-411
func NewPodSpec(serviceAccountName string, containers []core.Container, volumes []core.Volume, nodeSelector map[string]string, tolerations []core.Toleration) core.PodSpec {
	return core.PodSpec{
		Containers:         containers,
		ServiceAccountName: serviceAccountName,
		Volumes:            volumes,
		NodeSelector:       utils.EnsureLinuxNodeSelector(nodeSelector),
		Tolerations:        tolerations,
	}
}

//NewContainer stubs an instance of a Container
func NewContainer(containerName string, imageName string, pullPolicy core.PullPolicy, resources core.ResourceRequirements) core.Container {
	return core.Container{
		Name:            containerName,
		Image:           utils.GetComponentImage(imageName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}
