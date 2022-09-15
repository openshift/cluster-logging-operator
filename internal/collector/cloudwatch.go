package collector

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	v1 "k8s.io/api/core/v1"
	"path"
)

// Add volumes and env vars if output type is cloudwatch and role is found in the secret
func addWebIdentityForCloudwatch(collector *v1.Container, podSpec *v1.PodSpec, forwarderSpec logging.ClusterLogForwarderSpec, secrets map[string]*v1.Secret) {
	if secrets == nil {
		return
	}
	for _, o := range forwarderSpec.Outputs {
		// output secrets are keyed by output name
		secret := secrets[o.Name]
		if o.Type == logging.OutputTypeCloudwatch {
			if security.HasAwsRoleArnKey(secret) || security.HasAwsCredentialsKey(secret) {
				log.V(3).Info("Found sts key in secret")
				// Originally for fluentd and now for vector to use as well
				AddWebIdentityTokenVolumes(collector, podSpec)
				AddWebIdentityTokenEnvVars(collector, o, secret)
			}
		}
	}
}

// AddWebIdentityTokenVolumes Appends web identity volumes based on attributes of the secret and forwarder spec
func AddWebIdentityTokenVolumes(collector *v1.Container, podSpec *v1.PodSpec) {
	log.V(3).Info("Adding volumes for sts Cloudwatch")
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

// AddWebIdentityTokenEnvVars Appends web identity env vars based on attributes of the secret and forwarder spec
func AddWebIdentityTokenEnvVars(collector *v1.Container, output logging.OutputSpec, secret *v1.Secret) {
	// Necessary for vector to use sts
	// Also updated fluentd config to read from these as env vars
	log.V(3).Info("Adding env vars for sts Cloudwatch")
	collector.Env = append(collector.Env,
		v1.EnvVar{
			Name:  constants.AWSRegionEnvVarKey,
			Value: output.Cloudwatch.Region,
		},
		v1.EnvVar{
			Name:  constants.AWSRoleArnEnvVarKey,
			Value: cloudwatch.ParseRoleArn(secret),
		},
		v1.EnvVar{
			Name:  constants.AWSRoleSessionEnvVarKey,
			Value: constants.AWSRoleSessionName,
		},
		v1.EnvVar{
			Name:  constants.AWSWebIdentityTokenEnvVarKey,
			Value: path.Join(constants.AWSWebIdentityTokenMount, constants.AWSWebIdentityTokenFilePath),
		},
	)
}
