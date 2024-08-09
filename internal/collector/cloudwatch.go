package collector

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	v1 "k8s.io/api/core/v1"
)

// Add volumes and env vars if output type is cloudwatch and role is found in the secret
func addWebIdentityForCloudwatch(collector *v1.Container, forwarderSpec obs.ClusterLogForwarderSpec, secrets observability.Secrets) {
	if secrets == nil {
		return
	}
	for _, o := range forwarderSpec.Outputs {
		if o.Type == obs.OutputTypeCloudwatch && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.Type == obs.CloudwatchAuthTypeIAMRole {

			if roleARN := cloudwatch.ParseRoleArn(o.Cloudwatch.Authentication, secrets); roleARN != "" {
				tokenPath := common.ServiceAccountBasePath(constants.TokenKey)
				if o.Cloudwatch.Authentication.IAMRole.Token.From == obs.BearerTokenFromSecret {
					secret := o.Cloudwatch.Authentication.IAMRole.Token.Secret
					tokenPath = common.SecretPath(secret.Name, secret.Key)
				}

				AddWebIdentityTokenEnvVars(collector, o.Cloudwatch.Region, roleARN, tokenPath)
			}
		}
	}
}

// AddWebIdentityTokenEnvVars Appends web identity env vars based on attributes of the secret and forwarder spec
func AddWebIdentityTokenEnvVars(collector *v1.Container, region, roleARN, tokenPath string) {

	// Necessary for vector to use sts
	log.V(3).Info("Adding env vars for vector sts Cloudwatch")
	collector.Env = append(collector.Env,
		v1.EnvVar{
			Name:  constants.AWSRegionEnvVarKey,
			Value: region,
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
