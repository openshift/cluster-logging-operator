package fluentd

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
)

const (
	vectorConfigPath = "/etc/vector"
	vectorDataPath   = "/var/lib/vector"
)

func CollectorVisitor(collectorContainer *v1.Container, podSpec *v1.PodSpec) {
	collectorContainer.Env = append(collectorContainer.Env,
		v1.EnvVar{Name: "LOG", Value: "info"},
		v1.EnvVar{
			Name: "VECTOR_SELF_NODE_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1", FieldPath: "spec.nodeName",
				},
			},
		},
	)
	collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
		v1.VolumeMount{Name: common.ConfigVolumeName, ReadOnly: true, MountPath: vectorConfigPath},
		v1.VolumeMount{Name: common.DataDir, ReadOnly: false, MountPath: vectorDataPath},
	)
	podSpec.Volumes = append(podSpec.Volumes,
		v1.Volume{Name: common.ConfigVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorConfigSecretName, Optional: utils.GetBool(true)}}},
		v1.Volume{Name: common.DataDir, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: vectorDataPath}}},
	)
}
