package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("validating CloudWatch auth", func() {
	Context("#ValidateCloudWatchAuth", func() {

		It("should skip validation when not CloudWatch", func() {
			spec := obs.OutputSpec{
				Name: "myoutput",
				Type: obs.OutputTypeLoki,
			}
			Expect(ValidateCloudWatchAuth(spec)).To(BeEmpty())
		})

		DescribeTable("when using accessKey", func(auth *obs.CloudwatchAWSAccessKey, match types.GomegaMatcher) {
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
			cond := ValidateCloudWatchAuth(spec)
			Expect(cond).To(match)
		},
			Entry("should fail when not defined", nil, ContainElement(MatchRegexp(`AccessKey is nil`))),
			Entry("should fail when the KeyID is not defined", &obs.CloudwatchAWSAccessKey{
				KeySecret: &obs.SecretKey{},
			}, ContainElement(MatchRegexp(`KeyID.*`))),
			Entry("should fail when the KeySecret is not defined", &obs.CloudwatchAWSAccessKey{
				KeyID: &obs.SecretKey{},
			}, ContainElement(MatchRegexp(`KeySecret.*`))),
			Entry("should pass when all required fields are defined",
				&obs.CloudwatchAWSAccessKey{
					KeyID:     &obs.SecretKey{},
					KeySecret: &obs.SecretKey{},
				},
				BeEmpty()),
		)
		DescribeTable("when using IAMRole", func(auth *obs.CloudwatchIAMRole, match types.GomegaMatcher) {
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
			cond := ValidateCloudWatchAuth(spec)
			Expect(cond).To(match)
		},
			Entry("should fail when not defined", nil, ContainElement(MatchRegexp(`IAMRole is nil`))),
			Entry("should fail when the role_arn is not defined", &obs.CloudwatchIAMRole{
				Token: &obs.BearerToken{},
			}, ContainElement(MatchRegexp(`RoleARN.*`))),
			Entry("should pass when the role_arn is defined without a token", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
			}, BeEmpty()),
			Entry("should fail when the token is sourced from a secret without the secreting being defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
				},
			}, ContainElement(MatchRegexp(`Secret for token.*`))),
			Entry("should pass when the token is from a secret and role_arn are defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
				Token: &obs.BearerToken{
					From:   obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{},
				},
			}, BeEmpty()),
			Entry("should pass when the token is from a serviceaccount and role are defined", &obs.CloudwatchIAMRole{
				RoleARN: &obs.SecretKey{},
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromServiceAccountToken,
				},
			}, BeEmpty()),
		)
	})
})
