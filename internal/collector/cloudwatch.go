package collector

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	v1 "k8s.io/api/core/v1"
	"path"
)

// Add volumes and env vars if output type is cloudwatch and role is found in the secret
func addWebIdentityForCloudwatch(collector *v1.Container, podSpec *v1.PodSpec, forwarderSpec obs.ClusterLogForwarderSpec, secrets helpers.Secrets) {
	if secrets == nil {
		return
	}
	//for _, o := range forwarderSpec.Outputs {
	// TODO: fix me
	//if o.Type == obs.OutputTypeCloudwatch && o.Cloudwatch.Authentication != nil {
	//
	//	auth := o.Cloudwatch.Authentication
	//	if auth.Credentials != nil || auth.RoleARN != nil {
	//		roleARN := cloudwatch.ParseRoleArn(auth, secrets)
	//		AddWebIdentityTokenEnvVars(collector, o, roleARN)
	//	}
	//
	//	secret := secrets[o.Name]
	//	if security.HasAwsRoleArnKey(secret) || security.HasAwsCredentialsKey(secret) {
	//		log.V(3).Info("Found sts key in secret")
	//		if !security.HasAWSWebIdentityTokenFilePath(secret) { //assume legacy case to use SA token
	//			AddWebIdentityTokenVolumes(collector, podSpec)
	//		}
	//		AddWebIdentityTokenEnvVars(collector, o, secret)
	//		return
	//	}
	//}
	//}
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
								Path:     constants.TokenKey,
							},
						},
					},
				},
			},
		})

}

// AddWebIdentityTokenEnvVars Appends web identity env vars based on attributes of the secret and forwarder spec
func AddWebIdentityTokenEnvVars(collector *v1.Container, output obs.OutputSpec, roleARN string) {
	tokenPath := path.Join(constants.AWSWebIdentityTokenMount, constants.TokenKey)
	// TODO: fix me or delete me
	//if security.HasAWSWebIdentityTokenFilePath(secret) {
	//	tokenPath = common.SecretPath(secret.Name, constants.TokenKey)
	//}

	// Necessary for vector to use sts
	log.V(3).Info("Adding env vars for vector sts Cloudwatch")
	collector.Env = append(collector.Env,
		v1.EnvVar{
			Name:  constants.AWSRegionEnvVarKey,
			Value: output.Cloudwatch.Region,
		},
		v1.EnvVar{
			Name:  constants.AWSRoleArnEnvVarKey,
			Value: roleARN,
		},
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
