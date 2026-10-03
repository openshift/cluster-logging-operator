package vector

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

const initContainerName = "data-dir-init"

func CollectorVisitor(collectorContainer *corev1.Container, podSpec *corev1.PodSpec, resNames *factory.ForwarderResourceNames, namespace, logLevel string) {
	collectorContainer.Env = append(collectorContainer.Env,
		corev1.EnvVar{Name: "VECTOR_LOG", Value: logLevel},
		corev1.EnvVar{Name: "KUBERNETES_SERVICE_HOST", Value: "kubernetes.default.svc"},
		corev1.EnvVar{
			Name: "VECTOR_SELF_NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1", FieldPath: "spec.nodeName",
				},
			},
		},
	)

	dataPath := GetDataPath(namespace, resNames.ForwarderName)

	podSpec.InitContainers = append(podSpec.InitContainers, NewDataDirInitContainer(dataPath))

	collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
		corev1.VolumeMount{Name: common.ConfigVolumeName, ReadOnly: true, MountPath: vectorConfigPath},
		corev1.VolumeMount{Name: common.DataDir, ReadOnly: false, MountPath: dataPath},
		corev1.VolumeMount{Name: common.EntrypointVolumeName, ReadOnly: true, MountPath: entrypointValue, SubPath: RunVectorFile},
	)

	collectorContainer.Command = []string{"sh"}
	collectorContainer.Args = []string{entrypointValue}

	hostPathDirOrCreate := corev1.HostPathDirectoryOrCreate
	podSpec.Volumes = append(podSpec.Volumes,
		corev1.Volume{Name: common.ConfigVolumeName, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: resNames.ConfigMap}}}},
		corev1.Volume{Name: common.DataDir, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: dataPath, Type: &hostPathDirOrCreate}}},
		corev1.Volume{Name: common.EntrypointVolumeName, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: resNames.ConfigMap}}}},
	)
}

func PodLogExcludeLabel(o runtime.Object) {
	utils.AddLabels(runtime.Meta(o), map[string]string{"vector.dev/exclude": "true"})
}

// NewDataDirInitContainer creates an init container that prepares the collector's data directory.
// The init container runs as root with spc_t to create the directory and relabel it to
// container_file_t (which container_logwriter_t can write to). Group permissions (2770) grant
// Vector (UID 1000, GID 0) full read/write access.
func NewDataDirInitContainer(dataPath string) corev1.Container {
	return corev1.Container{
		Name:    initContainerName,
		Image:   utils.GetComponentImage(constants.VectorName),
		Command: []string{"/bin/sh", "-c"},
		Args: []string{
			"set -e; " +
				"mkdir -p \"$1\" && " +
				"chmod -R 2770 \"$1\" && " +
				"(chcon -R -t container_file_t \"$1\" || true)",
			"--",
			dataPath,
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: common.DataDir, MountPath: dataPath},
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.GetPtr(false),
			ReadOnlyRootFilesystem:   utils.GetPtr(true),
			SELinuxOptions: &corev1.SELinuxOptions{
				Type: common.SelinuxTypeSpc,
			},
			RunAsUser: utils.GetPtr[int64](0),
		},
	}
}
