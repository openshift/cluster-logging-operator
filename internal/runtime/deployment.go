package runtime

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentBuilder struct {
	Deployment *apps.Deployment
}

func NewDeploymentBuilder(ds *apps.Deployment) *DeploymentBuilder {
	return &DeploymentBuilder{
		Deployment: ds,
	}
}

func (builder *DeploymentBuilder) WithTemplateAnnotations(annotations map[string]string) *DeploymentBuilder {
	builder.Deployment.Spec.Template.Annotations = annotations
	return builder
}

func (builder *DeploymentBuilder) WithSelector(selector map[string]string) *DeploymentBuilder {
	builder.Deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: selector,
	}
	return builder
}

func (builder *DeploymentBuilder) WithTemplateLabels(labels map[string]string) *DeploymentBuilder {
	builder.Deployment.Spec.Template.Labels = labels
	return builder
}

func (builder *DeploymentBuilder) WithUpdateStrategy(updateStrategy apps.DeploymentStrategy) *DeploymentBuilder {
	builder.Deployment.Spec.Strategy = updateStrategy
	return builder
}

func (builder *DeploymentBuilder) WithPodSpec(podSpec core.PodSpec) *DeploymentBuilder {
	builder.Deployment.Spec.Template.Spec = podSpec
	return builder
}

func (builder *DeploymentBuilder) WithReplicas(replicas *int32) *DeploymentBuilder {
	builder.Deployment.Spec.Replicas = replicas
	return builder
}
