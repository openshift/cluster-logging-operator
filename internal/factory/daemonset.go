package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewDaemonSet stubs an instance of a daemonset
func NewDaemonSet(namespace, daemonsetName, instanceName, component, impl, maxUnavailable string, podSpec core.PodSpec, visitors ...func(o runtime.Object)) *apps.DaemonSet {
	selectors := runtime.Selectors(instanceName, component, impl)
	annotations := map[string]string{
		"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
	}
	intOrStringValue := intstr.Parse(maxUnavailable)
	strategy := apps.DaemonSetUpdateStrategy{
		Type: apps.RollingUpdateDaemonSetStrategyType,
		RollingUpdate: &apps.RollingUpdateDaemonSet{
			MaxUnavailable: &intOrStringValue,
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

// GetMaxUnavailableValue checks the framework options for the flag maxUnavailableRollout
// Default is 100%
func GetMaxUnavailableValue(op framework.Options) string {
	value, _ := utils.GetOption(op, framework.MaxUnavailableOption, "100%")
	if !observability.IsPercentOrWholeNumber(value) {
		value = "100%"
	}
	return value
}
