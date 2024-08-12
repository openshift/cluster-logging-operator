package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// NewDeployment stubs an instance of a deployment
func NewDeployment(namespace, deploymentName, component, impl string, replicas int32, podSpec core.PodSpec, visitors ...func(o runtime.Object)) *apps.Deployment {
	selectors := runtime.Selectors(deploymentName, component, impl)

	annotations := map[string]string{
		"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
	}

	dpl := runtime.NewDeployment(namespace, deploymentName, visitors...)
	runtime.NewDeploymentBuilder(dpl).WithTemplateAnnotations(annotations).
		WithTemplateLabels(dpl.Labels).
		WithSelector(selectors).
		WithPodSpec(podSpec).
		WithReplicas(utils.GetPtr(replicas))
	return dpl
}
