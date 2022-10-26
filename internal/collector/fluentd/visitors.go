package fluentd

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	certsVolumeName  = "certs"
	certsVolumePath  = "/etc/fluent/keys"
	configVolumePath = "/etc/fluent/configs.d/user"
	dataDir          = "/var/lib/fluentd"
	entrypointValue  = "/opt/app-root/src/run.sh"
)

var (
	DefaultMemory     = resource.MustParse("736Mi")
	DefaultCpuRequest = resource.MustParse("100m")
)

func CollectorVisitor(collectorContainer *v1.Container, podSpec *v1.PodSpec) {

	collectorContainer.Env = append(collectorContainer.Env,
		v1.EnvVar{
			Name: "NODE_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1", FieldPath: "spec.nodeName",
				},
			},
		},
		v1.EnvVar{Name: "RUBY_GC_HEAP_OLDOBJECT_LIMIT_FACTOR", Value: "0.9"},
	)
	collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
		v1.VolumeMount{Name: certsVolumeName, ReadOnly: true, MountPath: certsVolumePath},
		v1.VolumeMount{Name: common.ConfigVolumeName, ReadOnly: true, MountPath: configVolumePath},
		v1.VolumeMount{Name: common.DataDir, ReadOnly: false, MountPath: dataDir},
		v1.VolumeMount{Name: common.EntrypointVolumeName, ReadOnly: true, MountPath: entrypointValue, SubPath: "run.sh"},
	)

	podSpec.Volumes = append(podSpec.Volumes,
		v1.Volume{Name: certsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorName, Optional: utils.GetBool(true)}}},
		v1.Volume{Name: common.ConfigVolumeName, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: constants.CollectorName}}}},
		v1.Volume{Name: common.DataDir, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: dataDir}}},
		v1.Volume{Name: common.EntrypointVolumeName, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: constants.CollectorName}}}},
	)
}
