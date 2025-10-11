package outputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/aws"
	_ "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
)

const (
	ErrInvalidRoleARN       = "AWS RoleARN is invalid"
	ErrInvalidAssumeRoleARN = "AWS AssumeRole RoleARN is invalid"
)

// ValidateAwsAuth ensures auth role and assumeRole ARN's are valid
func ValidateAwsAuth(o obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	secrets := observability.Secrets(context.Secrets)
	if isRoleAuth, awsAuth := aws.OutputIsAwsRoleAuth(o); isRoleAuth {
		roleArn := aws.ParseRoleArn(awsAuth, secrets)
		if roleArn == "" {
			results = append(results, ErrInvalidRoleARN)
		}
	}
	// Additional validation for new assumeRole spec
	if isAssumeRole, assumeRoleSpec := aws.OutputIsAssumeRole(o); isAssumeRole {
		assumeRoleArn := aws.ParseAssumeRoleArn(assumeRoleSpec, secrets)
		if assumeRoleArn == "" {
			results = append(results, ErrInvalidAssumeRoleARN)
		}
	}
	return results
}
