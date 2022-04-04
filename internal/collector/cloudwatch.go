package collector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	v1 "k8s.io/api/core/v1"
)

// CloudwatchSecretWithRoleArnKey return true if secret has 'role_arn' key and output type is cloudwatch
func CloudwatchSecretWithRoleArnKey(secrets map[string]*v1.Secret, forwarderSpec logging.ClusterLogForwarderSpec) bool {
	if secrets == nil {
		return false
	}
	for _, o := range forwarderSpec.Outputs {
		secret := secrets[o.Name]
		if security.HasAwsRoleArnKey(secret) && o.Type == logging.OutputTypeCloudwatch {
			return true
		}
	}
	return false
}

func addVolumesForCloudwatch(collector *v1.Container, podSpec *v1.PodSpec, forwarderSpec logging.ClusterLogForwarderSpec, secrets map[string]*v1.Secret) {

	// Append any additional volumes based on attributes of the secret and forwarder spec
	if CloudwatchSecretWithRoleArnKey(secrets, forwarderSpec) {
		collector.VolumeMounts = append(collector.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.AWSWebIdentityTokenName,
				ReadOnly:  true,
				MountPath: constants.AWSWebIdentityTokenMount,
			})
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{
				Name: constants.AWSWebIdentityTokenName,
				VolumeSource: v1.VolumeSource{
					Projected: &v1.ProjectedVolumeSource{
						Sources: []v1.VolumeProjection{
							{
								ServiceAccountToken: &v1.ServiceAccountTokenProjection{
									Audience: "openshift",
									Path:     constants.AWSWebIdentityTokenFilePath,
								},
							},
						},
					},
				},
			})
	}
}
