package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// NewDeployment stubs an instance of a deployment
func NewDeployment(namespace, deploymentName, loggingComponent, component, impl string, podSpec core.PodSpec, visitors ...func(o runtime.Object)) *apps.Deployment {
	selectors := map[string]string{
		"provider":                        "openshift",
		"component":                       component,
		"logging-infra":                   loggingComponent,
		constants.CollectorDeploymentKind: constants.DeploymentType,
	}
	labels := map[string]string{
		"implementation": impl,
	}

	for k, v := range selectors {
		labels[k] = v
	}

	annotations := map[string]string{
		"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
	}

	// Number of replicas for the deployment
	replicas := int32(2)

	dpl := runtime.NewDeployment(namespace, deploymentName, visitors...)
	utils.AddLabels(dpl, labels)
	runtime.NewDeploymentBuilder(dpl).WithTemplateAnnotations(annotations).
		WithTemplateLabels(dpl.Labels).
		WithSelector(selectors).
		WithPodSpec(podSpec).
		WithReplicas(&replicas)
	return dpl
}
