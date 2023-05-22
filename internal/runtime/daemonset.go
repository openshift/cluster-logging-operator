package runtime

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSetBuilder struct {
	DaemonSet *apps.DaemonSet
}

func NewDaemonSetBuilder(ds *apps.DaemonSet) *DaemonSetBuilder {
	return &DaemonSetBuilder{
		DaemonSet: ds,
	}
}

func (builder *DaemonSetBuilder) WithTemplateAnnotations(annotations map[string]string) *DaemonSetBuilder {
	builder.DaemonSet.Spec.Template.Annotations = annotations
	return builder
}

func (builder *DaemonSetBuilder) WithSelector(selector map[string]string) *DaemonSetBuilder {
	builder.DaemonSet.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: selector,
	}
	return builder
}

func (builder *DaemonSetBuilder) WithTemplateLabels(labels map[string]string) *DaemonSetBuilder {
	builder.DaemonSet.Spec.Template.Labels = labels
	return builder
}

func (builder *DaemonSetBuilder) WithUpdateStrategy(updateStrategy apps.DaemonSetUpdateStrategy) *DaemonSetBuilder {
	builder.DaemonSet.Spec.UpdateStrategy = updateStrategy
	return builder
}

func (builder *DaemonSetBuilder) WithPodSpec(podSpec core.PodSpec) *DaemonSetBuilder {
	builder.DaemonSet.Spec.Template.Spec = podSpec
	return builder
}
