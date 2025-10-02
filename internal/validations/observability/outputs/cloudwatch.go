package outputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	_ "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
)

const (
	ErrInvalidRoleARN       = "CloudWatch RoleARN is invalid"
	ErrInvalidAssumeRoleARN = "CloudWatch AssumeRole RoleARN is invalid"
)

// ValidateCloudWatchAuth ensures auth role and assumeRole ARN's are valid
func ValidateCloudWatchAuth(o obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	secrets := observability.Secrets(context.Secrets)
	if isRoleAuth, awsAuth := cloudwatch.OutputIsRoleAuth(o); isRoleAuth {
		roleArn := cloudwatch.ParseRoleArn(awsAuth, secrets)
		if roleArn == "" {
			results = append(results, ErrInvalidRoleARN)
		}
	}
	// Additional validation for new assumeRole spec
	if isAssumeRole, assumeRoleSpec := cloudwatch.OutputIsAssumeRole(o); isAssumeRole {
		assumeRoleArn := cloudwatch.ParseAssumeRoleArn(assumeRoleSpec, secrets)
		if assumeRoleArn == "" {
			results = append(results, ErrInvalidAssumeRoleARN)
		}
	}
	return results
}
