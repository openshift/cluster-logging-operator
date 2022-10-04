package fluentd

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	vectorConfigPath = "/etc/vector"
	vectorDataPath   = "/var/lib/vector"
)

func CollectorVisitor(collectorContainer *corev1.Container, podSpec *corev1.PodSpec) {
	collectorContainer.Env = append(collectorContainer.Env,
		corev1.EnvVar{Name: "VECTOR_LOG", Value: "info"},
		corev1.EnvVar{
			Name: "VECTOR_SELF_NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1", FieldPath: "spec.nodeName",
				},
			},
		},
		corev1.EnvVar{Name: "OPENSSL_CONF", Value: fmt.Sprintf("%s/openssl.cnf", vectorConfigPath)},
	)
	collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
		corev1.VolumeMount{Name: common.ConfigVolumeName, ReadOnly: true, MountPath: vectorConfigPath},
		corev1.VolumeMount{Name: common.DataDir, ReadOnly: false, MountPath: vectorDataPath},
	)
	podSpec.Volumes = append(podSpec.Volumes,
		corev1.Volume{Name: common.ConfigVolumeName, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: constants.CollectorConfigSecretName, Optional: utils.GetBool(true)}}},
		corev1.Volume{Name: common.DataDir, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: vectorDataPath}}},
	)
}
