package utils

import (
	"fmt"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//NewDaemonSet stubs an instance of a daemonset
func NewDaemonSet(daemonsetName, namespace, loggingComponent, component string, podSpec core.PodSpec) *apps.DaemonSet {
	labels := map[string]string{
		"provider":      "openshift",
		"component":     component,
		"logging-infra": loggingComponent,
	}
	return &apps.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonsetName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: apps.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   daemonsetName,
					Labels: labels,
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: podSpec,
			},
			UpdateStrategy: apps.DaemonSetUpdateStrategy{
				Type:          apps.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &apps.RollingUpdateDaemonSet{},
			},
		},
	}
}

//GetDaemonSetList lists DS in namespace with given selector
func GetDaemonSetList(namespace, selector string) (*apps.DaemonSetList, error) {
	list := &apps.DaemonSetList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: apps.SchemeGroupVersion.String(),
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

//RemoveDaemonset with given name and namespace
func RemoveDaemonset(namespace, daemonsetName string) error {

	daemonset := NewDaemonSet(
		daemonsetName,
		namespace,
		daemonsetName,
		daemonsetName,
		core.PodSpec{},
	)

	err := sdk.Delete(daemonset)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v daemonset %v", daemonsetName, err)
	}

	return nil
}
