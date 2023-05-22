package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewDaemonSet stubs an instance of a daemonset
func NewDaemonSet(daemonsetName, namespace, loggingComponent, component, impl string, podSpec core.PodSpec, visitors ...func(o runtime.Object)) *apps.DaemonSet {
	selectors := map[string]string{
		"provider":      "openshift",
		"component":     component,
		"logging-infra": loggingComponent,
	}
	labels := map[string]string{

		"implementation": impl,
	}
	for k, v := range selectors {
		labels[k] = v
	}

	annotations := map[string]string{
		"scheduler.alpha.kubernetes.io/critical-pod": "",
		"target.workload.openshift.io/management":    `{"effect": "PreferredDuringScheduling"}`,
	}

	strategy := apps.DaemonSetUpdateStrategy{
		Type: apps.RollingUpdateDaemonSetStrategyType,
		RollingUpdate: &apps.RollingUpdateDaemonSet{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "100%",
			},
		},
	}
	ds := runtime.NewDaemonSet(namespace, daemonsetName, visitors...)
	utils.AddLabels(ds, labels) //todo
	runtime.NewDaemonSetBuilder(ds).WithTemplateAnnotations(annotations).
		WithTemplateLabels(ds.Labels).
		WithSelector(selectors).
		WithUpdateStrategy(strategy).
		WithPodSpec(podSpec)
	return ds
}
