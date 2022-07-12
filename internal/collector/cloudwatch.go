package collector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	v1 "k8s.io/api/core/v1"
)

// CloudwatchSecretWithRoleArn return true if output type is cloudwatch and secret has 'role_arn' or 'credentials' key
func CloudwatchSecretWithRoleArn(secrets map[string]*v1.Secret, forwarderSpec logging.ClusterLogForwarderSpec) bool {
	if secrets == nil {
		return false
	}
	for _, o := range forwarderSpec.Outputs {
		secret := secrets[o.Name]
		if o.Type == logging.OutputTypeCloudwatch {
			if security.HasAwsRoleArnKey(secret) || security.HasAwsCredentialsKey(secret) {
				return true
			}
		}
	}
	return false
}

// Append any additional volumes based on attributes of the secret and forwarder spec
func addVolumesForCloudwatch(collector *v1.Container, podSpec *v1.PodSpec, forwarderSpec logging.ClusterLogForwarderSpec, secrets map[string]*v1.Secret) {
	if CloudwatchSecretWithRoleArn(secrets, forwarderSpec) {
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
