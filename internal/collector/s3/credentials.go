package s3

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

// OutputIsS3RoleAuth identifies if `output.s3.Authentication.IamRole` exists and returns ref if so
func OutputIsS3RoleAuth(o obs.OutputSpec) (bool, *obs.AwsAuthentication) {
	if o.S3 != nil && o.S3.Authentication != nil && o.S3.Authentication.IamRole != nil {
		return true, o.S3.Authentication
	}
	return false, nil
}

func OutputIsS3AssumeRoleAuth(o obs.OutputSpec) (bool, *obs.AwsAssumeRole) {
	if o.S3 != nil && o.S3.Authentication != nil && o.S3.Authentication.AssumeRole != nil {
		return true, o.S3.Authentication.AssumeRole
	}
	return false, nil
}
