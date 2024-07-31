package cloudwatch

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	corev1 "k8s.io/api/core/v1"
)

func MapCloudwatch(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Cloudwatch {
	obsCw := &obs.Cloudwatch{
		URL: loggingOutSpec.URL,
	}

	// Map secret to authentication
	if secret != nil {
		obsCw.Authentication = &obs.CloudwatchAuthentication{}
		if security.HasAwsAccessKeyId(secret) && security.HasAwsSecretAccessKey(secret) {
			obsCw.Authentication.Type = obs.CloudwatchAuthTypeAccessKey
			obsCw.Authentication.AWSAccessKey = &obs.CloudwatchAWSAccessKey{
				KeyID: &obs.SecretReference{
					Key:        constants.AWSAccessKeyID,
					SecretName: secret.Name,
				},
				KeySecret: &obs.SecretReference{
					Key:        constants.AWSSecretAccessKey,
					SecretName: secret.Name,
				},
			}
		}

		if security.HasAwsRoleArnKey(secret) || security.HasAwsCredentialsKey(secret) {
			obsCw.Authentication.Type = obs.CloudwatchAuthTypeIAMRole

			// Determine if `role_arn` or `credentials` key is specified
			roleArnKey := constants.AWSWebIdentityRoleKey
			if security.HasAwsCredentialsKey(secret) {
				roleArnKey = constants.AWSCredentialsKey
			}

			obsCw.Authentication.IAMRole = &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretReference{
					Key:        roleArnKey,
					SecretName: secret.Name,
				},
			}
			if security.HasBearerTokenFileKey(secret) {
				obsCw.Authentication.IAMRole.Token = &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{
						Name: secret.Name,
						Key:  constants.BearerTokenFileKey,
					},
				}
			} else {
				obsCw.Authentication.IAMRole.Token = &obs.BearerToken{
					From: obs.BearerTokenFromServiceAccount,
				}
			}
		}
	}

	if loggingOutSpec.Tuning != nil {
		obsCw.Tuning = &obs.CloudwatchTuningSpec{
			Compression:          loggingOutSpec.Tuning.Compression,
			BaseOutputTuningSpec: *common.MapBaseOutputTuning(*loggingOutSpec.Tuning),
		}
	}

	loggingCw := loggingOutSpec.Cloudwatch
	if loggingCw == nil {
		return obsCw
	}

	obsCw.Region = loggingCw.Region

	// Group name
	groupBy := ""
	switch loggingCw.GroupBy {
	case logging.LogGroupByLogType:
		groupBy = ".log_type"
	case logging.LogGroupByNamespaceName:
		groupBy = ".kubernetes.namespace_name"
	case logging.LogGroupByNamespaceUUID:
		groupBy = ".kubernetes.namespace_uid"
	}
	groupPrefix := ""

	if loggingCw.GroupPrefix != nil {
		groupPrefix = *loggingCw.GroupPrefix
	} else {
		groupPrefix = `{.openshift.cluster_id||"none"}`
	}

	obsCw.GroupName = fmt.Sprintf(`%s.{%s||"none"}`, groupPrefix, groupBy)

	return obsCw
}
