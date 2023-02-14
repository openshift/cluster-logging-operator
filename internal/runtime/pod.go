package runtime

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

type PodBuilder struct {
	Pod *corev1.Pod
}

// PodBuilderVisitor provides the ability to manipulate the PodBuilder with
// custom logic
type PodBuilderVisitor func(builder *PodBuilder) error

func NewPodBuilder(pod *corev1.Pod) *PodBuilder {

	return &PodBuilder{
		Pod: pod,
	}
}

type ContainerBuilder struct {
	container  *corev1.Container
	podBuilder *PodBuilder
}

type InitContainerBuilder struct {
	initcontainer *corev1.Container
	podBuilder    *PodBuilder
}

func (builder *ContainerBuilder) End() *PodBuilder {
	builder.podBuilder.Pod.Spec.Containers = append(builder.podBuilder.Pod.Spec.Containers, *builder.container)
	return builder.podBuilder
}
func (builder *ContainerBuilder) Update() *PodBuilder {
	for i := range builder.podBuilder.Pod.Spec.Containers {
		if builder.podBuilder.Pod.Spec.Containers[i].Name == builder.container.Name {
			builder.podBuilder.Pod.Spec.Containers[i] = *(builder.container)
			return builder.podBuilder
		}
	}
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

func (builder *ContainerBuilder) ResourceRequirements(resources corev1.ResourceRequirements) *ContainerBuilder {
	builder.container.Resources = resources
	return builder
}

func (builder *ContainerBuilder) WithCmdArgs(cmdArgs []string) *ContainerBuilder {
	builder.container.Args = cmdArgs
	return builder
}

func (builder *ContainerBuilder) WithCmd(cmdAgrgs []string) *ContainerBuilder {
	builder.container.Command = []string{cmdAgrgs[0]}
	builder.container.Args = cmdAgrgs[1:]
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

func (builder *ContainerBuilder) WithPodSecurity() *ContainerBuilder {
	builder.container.SecurityContext = &corev1.SecurityContext{
		AllowPrivilegeEscalation: utils.GetBool(false),
		RunAsNonRoot:             utils.GetBool(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		SeccompProfile: &corev1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
	return builder
}

func (builder *ContainerBuilder) WithPrivilege() *ContainerBuilder {
	builder.container.SecurityContext = &corev1.SecurityContext{
		Privileged: utils.GetBool(true),
	}
	return builder
}
func (builder *ContainerBuilder) WithImagePullPolicy(policy corev1.PullPolicy) *ContainerBuilder {
	builder.container.ImagePullPolicy = policy
	return builder
}

func (builder *PodBuilder) AddContainer(name, image string) *ContainerBuilder {
	containerBuilder := ContainerBuilder{
		container: &corev1.Container{
			Name:            strings.ToLower(name),
			Image:           image,
			Env:             []corev1.EnvVar{},
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
		podBuilder: builder,
	}
	return &containerBuilder
}
func (builder *PodBuilder) AddConfigMapVolume(name, configMapName string) *PodBuilder {
	return builder.AddConfigMapVolumeWithPermissions(name, configMapName, utils.GetInt32(0644))
}

func (builder *PodBuilder) AddConfigMapVolumeWithPermissions(name, configMapName string, permissions *int32) *PodBuilder {
	builder.Pod.Spec.Volumes = append(builder.Pod.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
				DefaultMode: permissions,
			},
		},
	})
	return builder
}

func (builder *PodBuilder) AddSecretVolume(name, secretName string) *PodBuilder {
	builder.Pod.Spec.Volumes = append(builder.Pod.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	})
	return builder
}

func (builder *PodBuilder) GetContainer(name string) *ContainerBuilder {
	b := &ContainerBuilder{
		podBuilder: builder,
	}
	lowerCaseName := strings.ToLower(name)
	for i := range builder.Pod.Spec.Containers {
		if builder.Pod.Spec.Containers[i].Name == lowerCaseName {
			b.container = &builder.Pod.Spec.Containers[i]
			return b
		}
	}
	return b
}

func (builder *ContainerBuilder) AddRunAsUser(uid int64) *ContainerBuilder {
	builder.container.SecurityContext = &corev1.SecurityContext{
		RunAsUser: &uid,
	}
	return builder
}

func (builder *PodBuilder) AddAnnotation(key, value string) *PodBuilder {
	if builder.Pod.Annotations == nil {
		builder.Pod.Annotations = map[string]string{}
	}
	builder.Pod.Annotations[key] = value
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

func (builder *InitContainerBuilder) End() *PodBuilder {
	builder.podBuilder.Pod.Spec.InitContainers = append(builder.podBuilder.Pod.Spec.InitContainers, *builder.initcontainer)
	return builder.podBuilder
}

func (builder *InitContainerBuilder) WithCmdArgs(cmdAgrgs []string) *InitContainerBuilder {
	builder.initcontainer.Command = []string{cmdAgrgs[0]}
	builder.initcontainer.Args = cmdAgrgs[1:]
	return builder
}

func (builder *InitContainerBuilder) AddEnvVarFromEnvVarSource(name string, value string) *InitContainerBuilder {

	builder.initcontainer.Env = append(builder.initcontainer.Env, corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: value,
			},
		},
	})

	return builder
}

// added functions for support kafka in the functional pod
func (builder *InitContainerBuilder) AddVolumeMount(name, path, subPath string, readonly bool) *InitContainerBuilder {
	builder.initcontainer.VolumeMounts = append(builder.initcontainer.VolumeMounts, corev1.VolumeMount{
		Name:      name,
		ReadOnly:  readonly,
		MountPath: path,
		SubPath:   subPath,
	})
	return builder
}

func (builder *InitContainerBuilder) AddEnvVar(name, value string) *InitContainerBuilder {
	builder.initcontainer.Env = append(builder.initcontainer.Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
	return builder
}

func (builder *PodBuilder) AddInitContainer(name, image string) *InitContainerBuilder {
	containerBuilder := InitContainerBuilder{
		initcontainer: &corev1.Container{
			Name:            strings.ToLower(name),
			Image:           image,
			Env:             []corev1.EnvVar{},
			ImagePullPolicy: corev1.PullAlways,
		},
		podBuilder: builder,
	}
	return &containerBuilder
}

func (builder *PodBuilder) AddEmptyDirVolume(name string) *PodBuilder {
	builder.Pod.Spec.Volumes = append(builder.Pod.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
	return builder
}
func (builder *PodBuilder) AddHostPathVolume(name, path string) *PodBuilder {
	builder.Pod.Spec.Volumes = append(builder.Pod.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
			},
		},
	})
	return builder
}

func (builder *ContainerBuilder) AddContainerPort(name string, port int32) *ContainerBuilder {
	builder.container.Ports = append(builder.container.Ports, corev1.ContainerPort{Name: name, ContainerPort: port})
	return builder
}
