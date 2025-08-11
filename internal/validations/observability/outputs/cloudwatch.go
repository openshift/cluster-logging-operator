package outputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
)

const (
	RoleARNsOpt             = "roleARNs"
	ErrInvalidRoleARN       = "CloudWatch RoleARN is invalid"
	ErrInvalidAssumeRoleARN = "CloudWatch AssumeRole RoleARN is invalid"
)

func ValidateCloudWatchAuth(spec obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	secrets := observability.Secrets(context.Secrets)
	authSpec := spec.Cloudwatch.Authentication

	// Validate role ARN
	if authSpec.Type == obs.CloudwatchAuthTypeIAMRole {
		roleArn := cloudwatch.ParseRoleArn(authSpec, secrets)
		if roleArn == "" {
			results = append(results, ErrInvalidRoleARN)
		}
	}

	// Validate assume role ARN if specified
	if authSpec.AssumeRole != nil {
		assumeRoleArn := cloudwatch.ParseAssumeRoleArn(authSpec, secrets)
		if assumeRoleArn == "" {
			results = append(results, ErrInvalidAssumeRoleARN)
		}
	}

	return results
}
