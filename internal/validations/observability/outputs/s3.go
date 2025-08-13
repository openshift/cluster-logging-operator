package outputs

import (
	"regexp"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
)

const (
	ErrInvalidS3RoleARN       = "S3 RoleARN is invalid"
	ErrInvalidS3AssumeRoleARN = "S3 AssumeRole RoleARN is invalid"
)

func ValidateS3Auth(spec obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	secrets := observability.Secrets(context.Secrets)
	authSpec := spec.S3.Authentication

	// Validate role ARN
	if authSpec.Type == obs.S3AuthTypeIAMRole {
		roleArn := parseS3RoleArn(authSpec, secrets)
		if roleArn == "" {
			results = append(results, ErrInvalidS3RoleARN)
		}
	}

	// Validate assume role ARN if specified
	if authSpec.AssumeRole != nil {
		assumeRoleArn := parseS3AssumeRoleArn(authSpec, secrets)
		if assumeRoleArn == "" {
			results = append(results, ErrInvalidS3AssumeRoleARN)
		}
	}

	return results
}

// parseS3RoleArn search for matching valid ARN for S3 authentication
func parseS3RoleArn(auth *obs.S3Authentication, secrets observability.Secrets) string {
	if auth.Type == obs.S3AuthTypeIAMRole {
		roleArnString := secrets.AsString(&auth.IAMRole.RoleARN)

		if roleArnString != "" {
			reg := regexp.MustCompile(`(arn:aws(.*)?:(iam|sts)::\d{12}:role\/\S+)\s?`)
			roleArn := reg.FindStringSubmatch(roleArnString)
			if roleArn != nil {
				return roleArn[1] // the capturing group is index 1
			}
		}
	}
	return ""
}

// parseS3AssumeRoleArn search for matching valid assume role ARN for S3 authentication
func parseS3AssumeRoleArn(auth *obs.S3Authentication, secrets observability.Secrets) string {
	if auth.AssumeRole != nil {
		roleArnString := secrets.AsString(&auth.AssumeRole.RoleARN)

		if roleArnString != "" {
			reg := regexp.MustCompile(`(arn:aws(.*)?:(iam|sts)::\d{12}:role\/\S+)\s?`)
			roleArn := reg.FindStringSubmatch(roleArnString)
			if roleArn != nil {
				return roleArn[1] // the capturing group is index 1
			}
		}
	}
	return ""
}
