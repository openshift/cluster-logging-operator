package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//NewDaemonSet stubs an instance of a daemonset
func NewDaemonSet(daemonsetName, namespace, loggingComponent, component string, podSpec core.PodSpec) *apps.DaemonSet {
	labelSelectors := map[string]string{
		"provider":      "openshift",
		"component":     component,
		"logging-infra": loggingComponent,
	}
	labels := map[string]string{
		"app.kubernetes.io/name":       daemonsetName,
		"app.kubernetes.io/component":  constants.CollectorName,
		"app.kubernetes.io/created-by": constants.ClusterLoggingOperator,
		"app.kubernetes.io/managed-by": constants.ClusterLoggingOperator,
	}
	for k, v := range labelSelectors {
		labels[k] = v
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
				MatchLabels: labelSelectors,
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
				Type: apps.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &apps.RollingUpdateDaemonSet{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "100%",
					},
				},
			},
		},
	}
}
