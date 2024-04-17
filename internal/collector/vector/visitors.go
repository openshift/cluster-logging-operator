package vector

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

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
	collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
		corev1.VolumeMount{Name: common.ConfigVolumeName, ReadOnly: true, MountPath: vectorConfigPath},
		corev1.VolumeMount{Name: common.DataDir, ReadOnly: false, MountPath: dataPath},
		corev1.VolumeMount{Name: common.EntrypointVolumeName, ReadOnly: true, MountPath: entrypointValue, SubPath: RunVectorFile},
	)

	collectorContainer.Command = []string{"sh"}
	collectorContainer.Args = []string{entrypointValue}

	podSpec.Volumes = append(podSpec.Volumes,
		corev1.Volume{Name: common.ConfigVolumeName, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
			SecretName: resNames.ConfigMap,
			Items:      []corev1.KeyToPath{{Key: ConfigFile, Path: ConfigFile}},
			Optional:   utils.GetPtr(true),
		}}},
		corev1.Volume{Name: common.DataDir, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: dataPath}}},
		corev1.Volume{Name: common.EntrypointVolumeName, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
			SecretName: resNames.ConfigMap,
			Items:      []corev1.KeyToPath{{Key: RunVectorFile, Path: RunVectorFile}},
			Optional:   utils.GetPtr(true),
		}}},
	)
}

// PodLogExcludeLabel by default, the kubernetes_logs source will skip logs from the Pods that have a vector.dev/exclude: "true" label.
func PodLogExcludeLabel(o runtime.Object) {
	utils.AddLabels(runtime.Meta(o), map[string]string{"vector.dev/exclude": "true"})
}
