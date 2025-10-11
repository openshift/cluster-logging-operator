package aws

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/aws"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"strings"
)

type AwsAuth struct {
	KeyID           OptionalPair
	KeySecret       OptionalPair
	CredentialsPath OptionalPair
	Profile         OptionalPair
	AssumeRole      OptionalPair
	ExternalID      OptionalPair
	SessionName     OptionalPair
}

func NewAuth() AwsAuth {
	return AwsAuth{
		KeyID:           NewOptionalPair("auth.access_key_id", nil),
		KeySecret:       NewOptionalPair("auth.secret_access_key", nil),
		CredentialsPath: NewOptionalPair("auth.credentials_file", nil),
		Profile:         NewOptionalPair("auth.profile", nil),
		AssumeRole:      NewOptionalPair("auth.assume_role", nil),
		ExternalID:      NewOptionalPair("auth.external_id", nil),
		SessionName:     NewOptionalPair("auth.session_name", nil),
	}
}

func (a AwsAuth) Name() string {
	return "awsAuthTemplate"
}

func (a AwsAuth) Template() string {
	return `{{define "` + a.Name() + `" -}}
{{.KeyID}}
{{.KeySecret}}
{{.CredentialsPath}}
{{.Profile}}
{{.AssumeRole}}
{{.ExternalID}}
{{.SessionName}}
{{- end}}`
}

// AuthConfig returns the templated VRL containing cloudwatch and s3 auth configuration
func AuthConfig(outputName string, auth *obs.AwsAuthentication, options utils.Options, secrets observability.Secrets) framework.Element {
	authConfig := NewAuth()
	if auth == nil {
		return authConfig
	}
	switch auth.Type {
	case obs.AuthTypeAccessKey:
		authConfig.KeyID.Value = vectorhelpers.SecretFrom(&auth.AwsAccessKey.KeyId)
		authConfig.KeySecret.Value = vectorhelpers.SecretFrom(&auth.AwsAccessKey.KeySecret)
		// New assumeRole works with static keys as well
		if auth.AssumeRole != nil {
			authConfig.AssumeRole.Value = vectorhelpers.SecretFrom(&auth.AssumeRole.RoleARN)
			// Optional externalID and sessionName
			if hasExtID, extID := aws.AssumeRoleHasExternalId(auth.AssumeRole); hasExtID {
				authConfig.ExternalID.Value = extID
			}
			if auth.AssumeRole.SessionName != "" {
				authConfig.SessionName.Value = auth.AssumeRole.SessionName
			}
		}
	case obs.AuthTypeIAMRole:
		if forwarderName, found := utils.GetOption(options, framework.OptionForwarderName, ""); found {
			// For OIDC roles we mount a configMap containing a credentials file
			authConfig.CredentialsPath.Value = strings.Trim(vectorhelpers.ConfigPath(forwarderName+"-"+constants.AwsCredentialsConfigMapName, constants.AwsCredentialsKey), `"`)
			authConfig.Profile.Value = "output_" + outputName
		}
	}
	return authConfig
}
