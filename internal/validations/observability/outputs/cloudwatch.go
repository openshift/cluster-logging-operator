package outputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func ValidateCloudWatchAuth(spec obs.OutputSpec) (results []string) {
	if spec.Type != obs.OutputTypeCloudwatch {
		return results
	}
	authSpec := spec.Cloudwatch.Authentication
	if authSpec == nil {
		return []string{"auth missing"}
	}
	switch authSpec.Type {
	case obs.CloudwatchAuthTypeAccessKey:
		if authSpec.AWSAccessKey == nil {
			return []string{"AccessKey is nil"}
		}
		if authSpec.AWSAccessKey.KeySecret == nil {
			results = append(results, "KeySecret is missing")
		}
		if authSpec.AWSAccessKey.KeyID == nil {
			results = append(results, "KeyID is missing")
		}
	case obs.CloudwatchAuthTypeIAMRole:
		if authSpec.IAMRole == nil {
			return []string{"IAMRole is nil"}
		}
		if authSpec.IAMRole.RoleARN == nil {
			results = append(results, "RoleARN is missing")
		}
		if authSpec.IAMRole.Token != nil && authSpec.IAMRole.Token.From == obs.BearerTokenFromSecret && authSpec.IAMRole.Token.Secret == nil {
			results = append(results, "Secret for token is missing")
		}
	}
	return results
}
