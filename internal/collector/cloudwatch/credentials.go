package cloudwatch

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

// OutputIsCloudwatchRoleAuth identifies if `output.Cloudwatch.Authentication.IamRole` exists and returns ref if so
func OutputIsCloudwatchRoleAuth(o obs.OutputSpec) (bool, *obs.AwsAuthentication) {
	if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.IamRole != nil {
		return true, o.Cloudwatch.Authentication
	}
	return false, nil
}

func OutputIsCloudwatchAssumeRoleAuth(o obs.OutputSpec) (bool, *obs.AwsAssumeRole) {
	if o.Cloudwatch != nil && o.Cloudwatch.Authentication != nil && o.Cloudwatch.Authentication.AssumeRole != nil {
		return true, o.Cloudwatch.Authentication.AssumeRole
	}
	return false, nil
}
