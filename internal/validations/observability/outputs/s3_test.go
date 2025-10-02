package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("S3 validation", func() {
	var (
		secrets map[string]*corev1.Secret
		context internalcontext.ForwarderContext
	)

	BeforeEach(func() {
		secrets = map[string]*corev1.Secret{
			"valid-s3-role": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid-s3-role",
				},
				Data: map[string][]byte{
					"role-arn": []byte("arn:aws:iam::123456789012:role/my-role"),
				},
			},
			"invalid-s3-role": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-s3-role",
				},
				Data: map[string][]byte{
					"role-arn": []byte("invalid-arn"),
				},
			},
			"valid-assume-role": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid-assume-role",
				},
				Data: map[string][]byte{
					"role-arn": []byte("arn:aws:iam::123456789012:role/assume-role"),
				},
			},
			"invalid-assume-role": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "invalid-assume-role",
				},
				Data: map[string][]byte{
					"role-arn": []byte("invalid-assume-arn"),
				},
			},
		}
		context = internalcontext.ForwarderContext{
			Secrets: secrets,
		}
	})

	Context("IAM Role authentication", func() {
		It("should pass validation with valid role ARN", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "valid-s3-role",
							},
						},
					},
				},
			}

			results := ValidateS3Auth(output, context)
			Expect(results).To(BeEmpty())
		})

		It("should fail validation with invalid role ARN", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "invalid-s3-role",
							},
						},
					},
				},
			}

			results := ValidateS3Auth(output, context)
			Expect(results).To(ContainElement(ErrInvalidS3RoleARN))
		})
	})

	Context("Assume Role authentication", func() {
		It("should pass validation with valid assume role ARN", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "valid-s3-role",
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "valid-assume-role",
							},
						},
					},
				},
			}

			results := ValidateS3Auth(output, context)
			Expect(results).To(BeEmpty())
		})

		It("should fail validation with invalid assume role ARN", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "valid-s3-role",
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "invalid-assume-role",
							},
						},
					},
				},
			}

			results := ValidateS3Auth(output, context)
			Expect(results).To(ContainElement(ErrInvalidS3AssumeRoleARN))
		})

		It("should fail validation with both invalid role ARN and assume role ARN", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "invalid-s3-role",
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "invalid-assume-role",
							},
						},
					},
				},
			}

			results := ValidateS3Auth(output, context)
			Expect(results).To(ContainElement(ErrInvalidS3RoleARN))
			Expect(results).To(ContainElement(ErrInvalidS3AssumeRoleARN))
		})
	})

	Context("Access Key authentication", func() {
		It("should pass validation without role ARN checks", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeAccessKey,
						AWSAccessKey: &obs.S3AWSAccessKey{
							KeyId: obs.SecretReference{
								Key:        "aws_access_key_id",
								SecretName: "s3-access-key",
							},
							KeySecret: obs.SecretReference{
								Key:        "aws_secret_access_key",
								SecretName: "s3-access-key",
							},
						},
					},
				},
			}

			results := ValidateS3Auth(output, context)
			Expect(results).To(BeEmpty())
		})
	})
})
