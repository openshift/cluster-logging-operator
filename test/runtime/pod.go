package runtime

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

type PodBuilder struct {
	Pod *corev1.Pod
}

//PodBuilderVisitor provides the ability to manipulate the PodBuilder with
//custom logic
type PodBuilderVisitor func(builder *PodBuilder) error

func NewPodBuilder(pod *corev1.Pod) *PodBuilder {

	return &PodBuilder{
		Pod: pod,
	}
}

type ContainerBuilder struct {
	container  corev1.Container
	podBuilder *PodBuilder
}

func (builder *ContainerBuilder) End() *PodBuilder {
	builder.podBuilder.Pod.Spec.Containers = append(builder.podBuilder.Pod.Spec.Containers, builder.container)
	return builder.podBuilder
}

func (builder *ContainerBuilder) AddVolumeMount(name, path, subPath string, readonly bool) *ContainerBuilder {
	builder.container.VolumeMounts = append(builder.container.VolumeMounts, corev1.VolumeMount{
		Name:      name,
		ReadOnly:  readonly,
		MountPath: path,
		SubPath:   subPath,
	})
	return builder
}

func (builder *ContainerBuilder) WithCmdArgs(cmdAgrgs []string) *ContainerBuilder {
	builder.container.Args = cmdAgrgs
	return builder
}

func (builder *ContainerBuilder) WithCmd(cmdString string) *ContainerBuilder {
	cmd := strings.Split(cmdString, " ")
	builder.container.Command = []string{cmd[0]}
	builder.container.Args = cmd[1:]
	return builder
}

func (builder *ContainerBuilder) AddEnvVar(name, value string) *ContainerBuilder {
	builder.container.Env = append(builder.container.Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
	return builder
}

func (builder *ContainerBuilder) AddEnvVarFromFieldRef(name, fieldRef string) *ContainerBuilder {
	builder.container.Env = append(builder.container.Env, corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: fieldRef,
			},
		},
	})
	return builder
}

func (builder *ContainerBuilder) WithPrivilege() *ContainerBuilder {
	builder.container.SecurityContext = &corev1.SecurityContext{
		Privileged: utils.GetBool(true),
	}
	return builder
}

func (builder *PodBuilder) AddContainer(name, image string) *ContainerBuilder {
	containerBuilder := ContainerBuilder{
		container: corev1.Container{
			Name:  strings.ToLower(name),
			Image: image,
			Env:   []corev1.EnvVar{},
		},
		podBuilder: builder,
	}
	return &containerBuilder
}

func (builder *PodBuilder) AddConfigMapVolume(name, configMapName string) *PodBuilder {
	builder.Pod.Spec.Volumes = append(builder.Pod.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
	return builder
}

func (builder *ContainerBuilder) AddRunAsUser(uid int64) *ContainerBuilder {
	builder.container.SecurityContext = &corev1.SecurityContext{
		RunAsUser: &uid,
	}
	return builder
}

func (builder *PodBuilder) WithLabels(labels map[string]string) *PodBuilder {
	builder.Pod.Labels = labels
	return builder
}

func (builder *PodBuilder) AddLabels(labels map[string]string) *PodBuilder {
	for k, v := range labels {
		builder.Pod.Labels[k] = v
	}
	return builder
}
