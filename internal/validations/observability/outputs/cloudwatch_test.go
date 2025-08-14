package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validating CloudWatch auth", func() {
	Context("#ValidateCloudWatchAuth", func() {

		var (
			myRoleArn      = "arn:aws:iam::123456789012:role/my-role-to-assume"
			invalidRoleARN = "arn:aws:iam::123456789:role/other-role-to-assume"
			spec           = obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeIAMRole,
						IAMRole: &obs.CloudwatchIAMRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        constants.AWSCredentialsKey,
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			fooSecret = &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: "foo",
				},
				Data: map[string][]byte{
					constants.AWSCredentialsKey: []byte(myRoleArn),
					"invalidRoleARN":            []byte(invalidRoleARN),
				},
			}
			context = internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{spec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					fooSecret.Name: fooSecret,
				},
			}
		)

		It("should pass with valid role arn", func() {
			res := ValidateCloudWatchAuth(spec, context)
			Expect(res).To(BeEmpty())
		})

		It("should fail if Role ARN is invalid", func() {
			spec.Cloudwatch.Authentication.IAMRole.RoleARN.Key = "invalidRoleARN"
			res := ValidateCloudWatchAuth(spec, context)
			Expect(res).ToNot(BeEmpty())
			Expect(len(res)).To(BeEquivalentTo(1))
			Expect(res[0]).To(BeEquivalentTo(ErrInvalidRoleARN))
		})

		It("should pass with valid assume role arn (Role Auth)", func() {
			assumeRoleArn := "arn:aws:iam::987654321098:role/cross-account-role"

			// Create a new spec for this test
			testSpec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeIAMRole,
						IAMRole: &obs.CloudwatchIAMRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        constants.AWSCredentialsKey,
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        "assume_role_arn",
							},
						},
					},
				},
			}

			// Create a new secret for this test
			testSecret := &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: "foo",
				},
				Data: map[string][]byte{
					constants.AWSCredentialsKey: []byte(myRoleArn),
					"assume_role_arn":           []byte(assumeRoleArn),
				},
			}

			testContext := internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{testSpec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					testSecret.Name: testSecret,
				},
			}

			res := ValidateCloudWatchAuth(testSpec, testContext)
			Expect(res).To(BeEmpty())
		})

		It("should fail if assume role ARN is invalid (Role Auth)", func() {
			invalidAssumeRoleArn := "invalid-assume-role-arn"

			// Create a new spec for this test with IAM role and invalid assume role ARN
			testSpec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeIAMRole,
						IAMRole: &obs.CloudwatchIAMRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        constants.AWSCredentialsKey,
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        "invalid_assume_role_arn",
							},
						},
					},
				},
			}

			// Create a new secret for this test
			testSecret := &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: "foo",
				},
				Data: map[string][]byte{
					"key_id":                    []byte("AKIATEST"),
					"key_secret":                []byte("test-secret"),
					constants.AWSCredentialsKey: []byte("arn:aws:iam::123456789012:role/valid-role"),
					"invalid_assume_role_arn":   []byte(invalidAssumeRoleArn),
				},
			}

			testContext := internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{testSpec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					testSecret.Name: testSecret,
				},
			}

			res := ValidateCloudWatchAuth(testSpec, testContext)
			Expect(res).ToNot(BeEmpty())
			Expect(len(res)).To(BeEquivalentTo(1))
			Expect(res[0]).To(BeEquivalentTo(ErrInvalidAssumeRoleARN))
		})

		It("should fail if both role ARN and assume role ARN are invalid (Role Auth)", func() {
			invalidAssumeRoleArn := "invalid-assume-role-arn"

			// Create a new spec for this test with both invalid ARNs
			testSpec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeIAMRole,
						IAMRole: &obs.CloudwatchIAMRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        "invalidRoleARN",
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        "invalid_assume_role_arn",
							},
						},
					},
				},
			}

			// Create a new secret for this test
			testSecret := &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: "foo",
				},
				Data: map[string][]byte{
					"invalidRoleARN":          []byte(invalidRoleARN),
					"invalid_assume_role_arn": []byte(invalidAssumeRoleArn),
				},
			}

			testContext := internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{testSpec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					testSecret.Name: testSecret,
				},
			}

			res := ValidateCloudWatchAuth(testSpec, testContext)
			Expect(res).ToNot(BeEmpty())
			Expect(len(res)).To(BeEquivalentTo(2))
			Expect(res).To(ContainElement(ErrInvalidRoleARN))
			Expect(res).To(ContainElement(ErrInvalidAssumeRoleARN))
		})

		It("should pass with valid assume role arn (Key Auth)", func() {
			var (
				assumeRoleArn = "arn:aws:iam::987654321098:role/cross-account-role"
				secretName    = "foo"
				keyId         = "AKIAIOSFODNN7EXAMPLE"
				keySecret     = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
			)

			// Create a new spec
			testSpec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeAccessKey,
						AWSAccessKey: &obs.CloudwatchAWSAccessKey{
							KeyId: obs.SecretReference{
								Key:        constants.AWSAccessKeyID,
								SecretName: secretName,
							},
							KeySecret: obs.SecretReference{
								Key:        constants.AWSSecretAccessKey,
								SecretName: secretName,
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								SecretName: secretName,
								Key:        "assume_role_arn",
							},
						},
					},
				},
			}

			// Create a new secret for this test
			testSecret := &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: secretName,
				},
				Data: map[string][]byte{
					constants.AWSAccessKeyID:     []byte(keyId),
					constants.AWSSecretAccessKey: []byte(keySecret),
					"assume_role_arn":            []byte(assumeRoleArn),
				},
			}

			testContext := internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{testSpec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					testSecret.Name: testSecret,
				},
			}

			res := ValidateCloudWatchAuth(testSpec, testContext)
			Expect(res).To(BeEmpty())
		})

		It("should fail if assume role ARN is invalid (Key Auth)", func() {
			// Create a new spec for this test with IAM role and invalid assume role ARN
			secretName := "foo"
			testSpec := obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeAccessKey,
						AWSAccessKey: &obs.CloudwatchAWSAccessKey{
							KeyId: obs.SecretReference{
								Key:        constants.AWSAccessKeyID,
								SecretName: secretName,
							},
							KeySecret: obs.SecretReference{
								Key:        constants.AWSSecretAccessKey,
								SecretName: secretName,
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								SecretName: secretName,
								Key:        "invalid_assume_role_arn",
							},
						},
					},
				},
			}

			// Create a new secret
			testSecret := &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: secretName,
				},
				Data: map[string][]byte{
					constants.AWSAccessKeyID:     []byte("AKIATEST"),
					constants.AWSSecretAccessKey: []byte("test-secret"),
					"invalid_assume_role_arn":    []byte("invalid-assume-role-arn"),
				},
			}

			testContext := internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{testSpec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					testSecret.Name: testSecret,
				},
			}

			res := ValidateCloudWatchAuth(testSpec, testContext)
			Expect(res).ToNot(BeEmpty())
			Expect(len(res)).To(BeEquivalentTo(1))
			Expect(res[0]).To(BeEquivalentTo(ErrInvalidAssumeRoleARN))
		})
	})
})
