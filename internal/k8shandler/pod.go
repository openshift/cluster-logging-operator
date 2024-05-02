package k8shandler

import (
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// NewContainer stubs an instance of a Container
func NewContainer(containerName string, imageName string, pullPolicy core.PullPolicy, resources core.ResourceRequirements) core.Container {
	return factory.NewContainer(containerName, imageName, pullPolicy, resources)
}

// GetPodList for a given selector and namespace
func (clusterRequest *ClusterLoggingRequest) GetPodList(selector map[string]string) (*core.PodList, error) {
	list := &core.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: core.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(
		selector,
		list,
	)

	return list, err
}
