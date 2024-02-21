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
	container     *corev1.Container
	podBuilder    *PodBuilder
	initContainer bool
}

func (builder *ContainerBuilder) containers() *[]corev1.Container {
	if builder.initContainer {
		return &builder.podBuilder.Pod.Spec.InitContainers
	} else {
		return &builder.podBuilder.Pod.Spec.Containers
	}
}

func (builder *ContainerBuilder) End() *PodBuilder {
	containers := builder.containers()
	*containers = append(*containers, *builder.container)
	return builder.podBuilder
}

func (builder *ContainerBuilder) Update() *PodBuilder {
	containers := builder.containers()
	for i := range *containers {
		if (*containers)[i].Name == builder.container.Name {
			(*containers)[i] = *(builder.container)
			break
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
		AllowPrivilegeEscalation: utils.GetPtr(false),
		RunAsNonRoot:             utils.GetPtr(true),
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
		Privileged: utils.GetPtr(true),
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
	return builder.AddConfigMapVolumeWithPermissions(name, configMapName, utils.GetPtr[int32](0644))
}

func (builder *PodBuilder) AddConfigMapVolumeWithPermissions(name, configMapName string, permissions *int32) *PodBuilder {
	return builder.AddConfigMapWith(name, configMapName, func(volume corev1.Volume) corev1.Volume {
		volume.VolumeSource.ConfigMap.DefaultMode = permissions
		return volume
	})
}

func (builder *PodBuilder) AddConfigMapWith(name, configMapName string, handler func(corev1.Volume) corev1.Volume) *PodBuilder {
	volume := handler(corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
	builder.Pod.Spec.Volumes = append(builder.Pod.Spec.Volumes, volume)
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

func (builder *ContainerBuilder) WithImage(image string) *ContainerBuilder {
	builder.container.Image = image
	return builder
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

func (builder *PodBuilder) AddInitContainer(name, image string) *ContainerBuilder {
	containerBuilder := ContainerBuilder{
		container: &corev1.Container{
			Name:            strings.ToLower(name),
			Image:           image,
			Env:             []corev1.EnvVar{},
			ImagePullPolicy: corev1.PullAlways,
		},
		podBuilder:    builder,
		initContainer: true,
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

func (builder *PodBuilder) AddHostAlias(hostAlias corev1.HostAlias) *PodBuilder {
	builder.Pod.Spec.HostAliases = append(builder.Pod.Spec.HostAliases, hostAlias)
	return builder
}
