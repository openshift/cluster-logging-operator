package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("validating CloudWatch auth", func() {
	var (
		secrets map[string]*corev1.Secret
	)
	Context("#ValidateCloudWatchAuth", func() {

		It("should skip validation when not CloudWatch", func() {
			spec := obs.OutputSpec{
				Name: "myoutput",
				Type: obs.OutputTypeLoki,
			}
			Expect(ValidateCloudWatchAuth(spec, secrets)).To(BeEmpty())
		})

		DescribeTable("when using accessKey", func(auth *obs.CloudwatchAWSAccessKey, secrets map[string]*corev1.Secret, match types.GomegaMatcher) {
			spec := obs.OutputSpec{
				Name: "myoutput",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type:         obs.CloudwatchAuthTypeAccessKey,
						AWSAccessKey: auth,
					},
				},
			}
			cond := ValidateCloudWatchAuth(spec, secrets)
			Expect(cond).To(match)
		},
			Entry("should fail when not defined", nil, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" requires auth configuration`)),
			Entry("should fail when the KeyID is not defined", &obs.CloudwatchAWSAccessKey{
				KeySecret: &obs.SecretKey{},
			}, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" auth requires a KeyID`)),
			Entry("should fail when the KeySecret is not defined", &obs.CloudwatchAWSAccessKey{
				KeyID: &obs.SecretKey{},
			}, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" auth requires a KeySecret`)),
			Entry("should pass when all required fields are defined",
				&obs.CloudwatchAWSAccessKey{
					KeyID:     &obs.SecretKey{},
					KeySecret: &obs.SecretKey{},
				},
				secrets, BeEmpty()),
		)
		DescribeTable("when using IAMRole", func(auth *obs.CloudwatchIAMRole, secrets map[string]*corev1.Secret, match types.GomegaMatcher) {
			spec := obs.OutputSpec{
				Name: "myoutput",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type:    obs.CloudwatchAuthTypeIAMRole,
						IAMRole: auth,
					},
				},
			}
			cond := ValidateCloudWatchAuth(spec, secrets)
			Expect(cond).To(match)
		},
			Entry("should fail when not defined", nil, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" requires auth configuration`)),
			Entry("should fail when the role_arn is not defined", &obs.CloudwatchIAMRole{
				Token: &obs.BearerToken{},
			}, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" auth requires a RoleARN`)),
			Entry("should fail when the token is not defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
			}, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" auth requires a token`)),
			Entry("should fail when the token is sourced from a secret without the secreting being defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
				},
			}, secrets, HaveCondition(obs.ValidationCondition, true, obs.ReasonMissingSpec, `".*" auth from "secret" requires a secret`)),
			Entry("should pass when the token secret and role_arn are defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
				Token: &obs.BearerToken{
					From:   obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{},
				},
			}, secrets, BeEmpty()),
			Entry("should pass when the token serviceaccount and role are defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromServiceAccountToken,
				},
			}, secrets, BeEmpty()),
		)
	})
})
