package runtime

import (
	corev1 "k8s.io/api/core/v1"
)

type ConfigMapBuilder struct {
	ConfigMap *corev1.ConfigMap
}

func NewConfigMapBuilder(cm *corev1.ConfigMap) *ConfigMapBuilder {
	return &ConfigMapBuilder{
		ConfigMap: cm,
	}
}
func (builder *ConfigMapBuilder) Add(key, value string) *ConfigMapBuilder {
	builder.ConfigMap.Data[key] = value
	return builder
}
func (builder *ConfigMapBuilder) AddLabel(key, value string) *ConfigMapBuilder {
	if builder.ConfigMap.Labels == nil {
		builder.ConfigMap.Labels = map[string]string{}
	}
	builder.ConfigMap.Labels[key] = value
	return builder
}
func (builder *ConfigMapBuilder) AddAnnotation(key, value string) *ConfigMapBuilder {
	if builder.ConfigMap.Annotations == nil {
		builder.ConfigMap.Annotations = map[string]string{}
	}
	builder.ConfigMap.Annotations[key] = value
	return builder
}
