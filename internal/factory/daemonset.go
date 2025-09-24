package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewDaemonSet stubs an instance of a daemonset
func NewDaemonSet(namespace, daemonsetName, instanceName, component, impl string, maxUnavailable intstr.IntOrString, podSpec core.PodSpec, visitors ...func(o runtime.Object)) *apps.DaemonSet {
	selectors := runtime.Selectors(instanceName, component, impl)
	annotations := map[string]string{
		"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
	}
	strategy := apps.DaemonSetUpdateStrategy{
		Type: apps.RollingUpdateDaemonSetStrategyType,
		RollingUpdate: &apps.RollingUpdateDaemonSet{
			MaxUnavailable: &maxUnavailable,
		},
	}
	ds := runtime.NewDaemonSet(namespace, daemonsetName, visitors...)
	runtime.NewDaemonSetBuilder(ds).WithTemplateAnnotations(annotations).
		WithTemplateLabels(ds.Labels).
		WithSelector(selectors).
		WithUpdateStrategy(strategy).
		WithPodSpec(podSpec)
	return ds
}
