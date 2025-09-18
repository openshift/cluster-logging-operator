package outputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"regexp"
)

const (
	ErrInvalidRoleARN       = "CloudWatch RoleARN is invalid"
	ErrInvalidAssumeRoleARN = "CloudWatch AssumeRole RoleARN is invalid"
)

// ValidateCloudWatchAuth ensures auth role and assumeRole ARN's are valid
func ValidateCloudWatchAuth(spec obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	secrets := observability.Secrets(context.Secrets)
	authSpec := spec.Cloudwatch.Authentication

	switch authSpec.Type {
	case obs.CloudwatchAuthTypeIAMRole:
		roleArn := ParseRoleArn(authSpec, secrets)
		if roleArn == "" {
			results = append(results, ErrInvalidRoleARN)
		}
		if authSpec.IAMRole.AssumeRole != nil {
			assumeRoleArn := ParseAssumeRoleArn(authSpec, secrets)
			if assumeRoleArn == "" {
				results = append(results, ErrInvalidAssumeRoleARN)
			}
		}
	case obs.CloudwatchAuthTypeAccessKey:
		if authSpec.AWSAccessKey.AssumeRole != nil {
			assumeRoleArn := ParseAssumeRoleArn(authSpec, secrets)
			if assumeRoleArn == "" {
				results = append(results, ErrInvalidAssumeRoleARN)
			}
		}
	}

	return results
}

// ParseRoleArn search for valid AWS arn
func ParseRoleArn(auth *obs.CloudwatchAuthentication, secrets observability.Secrets) string {
	var roleString string
	if auth.Type == obs.CloudwatchAuthTypeIAMRole {
		roleString = secrets.AsString(&auth.IAMRole.RoleARN)
	}
	return findSubstring(roleString)
}

// ParseAssumeRoleArn search for valid AWS assumeRole arn
func ParseAssumeRoleArn(auth *obs.CloudwatchAuthentication, secrets observability.Secrets) string {
	var roleString string
	if auth.IAMRole != nil && auth.IAMRole.AssumeRole != nil {
		roleString = secrets.AsString(&auth.IAMRole.AssumeRole.RoleARN)
	}
	if auth.AWSAccessKey != nil && auth.AWSAccessKey.AssumeRole != nil {
		roleString = secrets.AsString(&auth.AWSAccessKey.AssumeRole.RoleARN)
	}
	return findSubstring(roleString)
}

// findSubstring matches regex on a valid AWS role arn
func findSubstring(roleString string) string {
	if roleString != "" {
		reg := regexp.MustCompile(`(arn:aws(.*)?:(iam|sts)::\d{12}:role\/\S+)\s?`)
		roleArn := reg.FindStringSubmatch(roleString)
		if roleArn != nil {
			return roleArn[1] // the capturing group is index 1
		}
	}
	return ""
}
