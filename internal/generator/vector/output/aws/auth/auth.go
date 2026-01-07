package auth

import (
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/aws"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func New(outputName string, auth *obs.AwsAuthentication, options utils.Options) (a *sinks.AwsAuth) {
	if auth == nil {
		return nil
	}
	a = &sinks.AwsAuth{}
	switch auth.Type {
	case obs.AwsAuthTypeAccessKey:
		a.AccessKeyId = vectorhelpers.SecretFrom(&auth.AwsAccessKey.KeyId)
		a.SecretAccessKey = vectorhelpers.SecretFrom(&auth.AwsAccessKey.KeySecret)
		// New assumeRole works with static keys as well
		if auth.AssumeRole != nil {
			a.AssumeRole = vectorhelpers.SecretFrom(&auth.AssumeRole.RoleARN)
			// Optional externalID and sessionName
			if hasExtID, extID := aws.AssumeRoleHasExternalId(auth.AssumeRole); hasExtID {
				a.ExternalId = extID
			}
			if auth.AssumeRole.SessionName != "" {
				a.SessionName = auth.AssumeRole.SessionName
			}
		}
	case obs.AwsAuthTypeIAMRole:
		if forwarderName, found := utils.GetOption(options, framework.OptionForwarderName, ""); found {
			// For OIDC roles we mount a configMap containing a credentials file
			a.CredentialsFile = strings.Trim(vectorhelpers.ConfigPath(forwarderName+"-"+constants.AwsCredentialsConfigMapName, constants.AwsCredentialsKey), `"`)
			a.Profile = "output_" + outputName
		}
	}
	return a
}
