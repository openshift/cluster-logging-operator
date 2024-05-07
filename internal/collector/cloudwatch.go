package collector

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	v1 "k8s.io/api/core/v1"
	"path"
)

// Add volumes and env vars if output type is cloudwatch and role is found in the secret
func addWebIdentityForCloudwatch(collector *v1.Container, podSpec *v1.PodSpec, forwarderSpec logging.ClusterLogForwarderSpec, secrets map[string]*v1.Secret, collectorType logging.LogCollectionType) {
	if secrets == nil {
		return
	}
	for _, o := range forwarderSpec.Outputs {
		if o.Type == logging.OutputTypeCloudwatch {
			secret := secrets[o.Name]
			if security.HasAwsRoleArnKey(secret) || security.HasAwsCredentialsKey(secret) {
				log.V(3).Info("Found sts key in secret")
				if !security.HasAWSWebIdentityTokenFilePath(secret) { //assume legacy case to use SA token
					AddWebIdentityTokenVolumes(collector, podSpec)
				}
				// LOG-4084 fluentd no longer setting env vars
				if collectorType == logging.LogCollectionTypeVector {
					log.V(3).Info("Found vector collector")
					AddWebIdentityTokenEnvVars(collector, o, secret)
				}
				return
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
	tokenPath := path.Join(constants.AWSWebIdentityTokenMount, constants.AWSWebIdentityTokenFilePath)
	if security.HasAWSWebIdentityTokenFilePath(secret) {
		tokenPath = path.Join(OutputSecretPath(secret.Name), constants.AWSWebIdentityTokenFilePath)
	}

	// Necessary for vector to use sts
	log.V(3).Info("Adding env vars for vector sts Cloudwatch")
	collector.Env = append(collector.Env,
		v1.EnvVar{
			Name:  constants.AWSRegionEnvVarKey,
			Value: output.Cloudwatch.Region,
		},
		// TODO: FIX ME
		//v1.EnvVar{
		//	Name:  constants.AWSRoleArnEnvVarKey,
		//	Value: cloudwatch.ParseRoleArn(secret),
		//},
		v1.EnvVar{
			Name:  constants.AWSRoleSessionEnvVarKey,
			Value: constants.AWSRoleSessionName,
		},
		v1.EnvVar{
			Name:  constants.AWSWebIdentityTokenEnvVarKey,
			Value: tokenPath,
		},
	)
}
