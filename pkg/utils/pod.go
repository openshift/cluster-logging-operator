package utils

import (
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewPodSpec is a constructor to instansiate a new PodSpec
func NewPodSpec(serviceAccountName string, containers []core.Container, volumes []core.Volume, nodeSelector map[string]string) core.PodSpec {
	return core.PodSpec{
		Containers:         containers,
		ServiceAccountName: serviceAccountName,
		Volumes:            volumes,
		NodeSelector:       nodeSelector,
	}
}

//NewContainer stubs an instance of a Container
func NewContainer(containerName string, pullPolicy core.PullPolicy, resources core.ResourceRequirements) core.Container {
	return core.Container{
		Name:            containerName,
		Image:           GetComponentImage(containerName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}

//GetPodList for a given selector and namespace
func GetPodList(namespace, selector string) (*core.PodList, error) {
	list := &core.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: core.SchemeGroupVersion.String(),
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}
