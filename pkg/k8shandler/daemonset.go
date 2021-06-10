package k8shandler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
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
						"target.workload.openshift.io/management":    `{"effect": "PreferredDuringScheduling"}`,
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

//RemoveDaemonset with given name and namespace
func (clusterRequest *ClusterLoggingRequest) RemoveDaemonset(daemonsetName string) error {

	daemonset := NewDaemonSet(
		daemonsetName,
		clusterRequest.Cluster.Namespace,
		daemonsetName,
		daemonsetName,
		core.PodSpec{},
	)

	err := clusterRequest.Delete(daemonset)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting %v daemonset %v", daemonsetName, err)
	}

	return nil
}
