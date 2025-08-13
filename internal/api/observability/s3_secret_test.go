package observability

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("S3 secret handling", func() {
	Context("SecretReferences", func() {
		It("should return access key secrets for S3 with access key authentication", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeS3,
				S3: &obsv1.S3{
					Authentication: &obsv1.S3Authentication{
						Type: obsv1.S3AuthTypeAccessKey,
						AWSAccessKey: &obsv1.S3AWSAccessKey{
							KeyId: obsv1.SecretReference{
								SecretName: "s3-secrets",
								Key:        "access-key-id",
							},
							KeySecret: obsv1.SecretReference{
								SecretName: "s3-secrets",
								Key:        "secret-access-key",
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(2))
			Expect(refs[0].SecretName).To(Equal("s3-secrets"))
			Expect(refs[0].Key).To(Equal("access-key-id"))
			Expect(refs[1].SecretName).To(Equal("s3-secrets"))
			Expect(refs[1].Key).To(Equal("secret-access-key"))
		})

		It("should return IAM role secrets for S3 with IAM role authentication", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeS3,
				S3: &obsv1.S3{
					Authentication: &obsv1.S3Authentication{
						Type: obsv1.S3AuthTypeIAMRole,
						IAMRole: &obsv1.S3IAMRole{
							RoleARN: obsv1.SecretReference{
								SecretName: "s3-role",
								Key:        "role-arn",
							},
							Token: obsv1.BearerToken{
								From: obsv1.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(1))
			Expect(refs[0].SecretName).To(Equal("s3-role"))
			Expect(refs[0].Key).To(Equal("role-arn"))
		})

		It("should return assume role secrets for S3 with cross-account access", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeS3,
				S3: &obsv1.S3{
					Authentication: &obsv1.S3Authentication{
						Type: obsv1.S3AuthTypeIAMRole,
						IAMRole: &obsv1.S3IAMRole{
							RoleARN: obsv1.SecretReference{
								SecretName: "s3-base-role",
								Key:        "role-arn",
							},
							Token: obsv1.BearerToken{
								From: obsv1.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obsv1.CloudwatchAssumeRole{
							RoleARN: obsv1.SecretReference{
								SecretName: "s3-assume-role",
								Key:        "role-arn",
							},
							ExternalID: &obsv1.SecretReference{
								SecretName: "s3-assume-role",
								Key:        "external-id",
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(3))
			Expect(refs[0].SecretName).To(Equal("s3-base-role"))
			Expect(refs[0].Key).To(Equal("role-arn"))
			Expect(refs[1].SecretName).To(Equal("s3-assume-role"))
			Expect(refs[1].Key).To(Equal("role-arn"))
			Expect(refs[2].SecretName).To(Equal("s3-assume-role"))
			Expect(refs[2].Key).To(Equal("external-id"))
		})
	})

	Context("NeedServiceAccountToken", func() {
		It("should return true for S3 with IAM role authentication using service account token", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeS3,
					S3: &obsv1.S3{
						Authentication: &obsv1.S3Authentication{
							Type: obsv1.S3AuthTypeIAMRole,
							IAMRole: &obsv1.S3IAMRole{
								Token: obsv1.BearerToken{
									From: obsv1.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeTrue())
		})

		It("should return false for S3 with access key authentication", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeS3,
					S3: &obsv1.S3{
						Authentication: &obsv1.S3Authentication{
							Type: obsv1.S3AuthTypeAccessKey,
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeFalse())
		})
	})
})
