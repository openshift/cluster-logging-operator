package s3_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/s3"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating vector config for s3 output", func() {

	const (
		roleArn    = "arn:aws:iam::123456789012:role/my-role-to-assume"
		secretName = "s3-secret"
	)
	Context("#New", func() {

		const (
			keyId                 = "AKIAIOSFODNN7EXAMPLE"
			keySecret             = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
			secretName            = "s3-secret"
			secretNameWebIdentity = "s3-secret-webidentity"
			secretWithCredentials = "secretwithcredentials"
			secretWithRole        = "s3-role-secret"
		)

		var (
			initOutput = func() obs.OutputSpec {
				return obs.OutputSpec{
					Type: obs.OutputTypeS3,
					Name: "s3",
					S3: &obs.S3{
						Region: "us-east-test",
						Bucket: "my-test-bucket",
						Authentication: &obs.AwsAuthentication{
							Type: obs.AwsAuthTypeAccessKey,
							AwsAccessKey: &obs.AwsAccessKey{
								KeyId: obs.SecretReference{
									Key:        constants.AwsAccessKeyID,
									SecretName: secretName,
								},
								KeySecret: obs.SecretReference{
									Key:        constants.AwsSecretAccessKey,
									SecretName: secretName,
								},
							},
						},
					},
				}
			}

			secrets = map[string]*corev1.Secret{
				secretName: {
					Data: map[string][]byte{
						constants.AwsAccessKeyID:     []byte(keyId),
						constants.AwsSecretAccessKey: []byte(keySecret),
					},
				},
				secretWithCredentials: {
					Data: map[string][]byte{
						constants.AwsCredentialsKey:  []byte("[default]\nrole_arn = " + roleArn + "\nweb_identity_token_file = /var/run/secrets/token"),
						constants.ClientPrivateKey:   []byte("-- key-- "),
						constants.TrustedCABundleKey: []byte("-- ca-bundle -- "),
					},
				},
				secretWithRole: {
					Data: map[string][]byte{
						"my_role": []byte(roleArn),
					},
				},
				secretNameWebIdentity: {
					Data: map[string][]byte{
						constants.AwsWebIdentityRoleKey: []byte("arn:aws:iam::123456789012:role/initial-role"),
						"assume_role_arn":               []byte("arn:aws:iam::987654321098:role/cross-account-role"),
						"external_id":                   []byte("unique-web-external-id"),
					},
				},
			}
		)

		DescribeTable("should generate valid config", func(visit func(spec *obs.OutputSpec), op framework.Options, expFile string) {
			exp, err := testFiles.ReadFile(expFile)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
			}
			outputSpec := initOutput()
			if visit != nil {
				visit(&outputSpec)
			}
			op[framework.OptionForwarderName] = "my-forwarder"
			id, sink, transforms := s3.New(outputSpec.Name, adapters.NewOutput(outputSpec), []string{"s3-forward"}, secrets, op)
			Expect(exp).To(EqualConfigFrom(api.NewConfig(func(c *api.Config) {
				c.Sinks[id] = sink
				c.AddTransforms(transforms)
			})))
		},

			Entry("when a role_arn is provided directly", func(spec *obs.OutputSpec) {
				spec.S3.KeyPrefix = "app-{.log_type||\"missing\"}"
				spec.S3.Authentication = &obs.AwsAuthentication{
					Type: obs.AwsAuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        "my_role",
							SecretName: secretWithRole,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}
			}, framework.NoOptions, "files/s3_with_aws_credentials.toml"),

			Entry("when a role_arn is provided by ccoctl", func(spec *obs.OutputSpec) {
				spec.S3.KeyPrefix = "app-{.log_type||\"missing\"}"
				spec.S3.Authentication = &obs.AwsAuthentication{
					Type: obs.AwsAuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        constants.AwsCredentialsKey,
							SecretName: secretWithCredentials,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}
			}, framework.NoOptions, "files/s3_with_aws_credentials.toml"),

			Entry("when an assume_role is specified with accessKey auth", func(spec *obs.OutputSpec) {
				spec.S3.KeyPrefix = "app-{.log_type||\"missing\"}"
				spec.S3.Authentication = &obs.AwsAuthentication{
					Type: obs.AwsAuthTypeAccessKey,
					AwsAccessKey: &obs.AwsAccessKey{
						KeyId: obs.SecretReference{
							Key:        constants.AwsAccessKeyID,
							SecretName: secretName,
						},
						KeySecret: obs.SecretReference{
							Key:        constants.AwsSecretAccessKey,
							SecretName: secretName,
						},
					},
					AssumeRole: &obs.AwsAssumeRole{
						RoleARN: obs.SecretReference{
							Key:        "my_role",
							SecretName: secretWithRole,
						},
						ExternalID: "unique-external-id",
					},
				}
			}, framework.NoOptions, "files/s3_key_auth_and_assume_role.toml"),
			Entry("when an assume_role is specified with iamRole", func(spec *obs.OutputSpec) {
				spec.S3 = &obs.S3{
					Region:    "us-east-test",
					KeyPrefix: "app-{.log_type||\"missing\"}",
					Bucket:    "my-test-bucket",
					Authentication: &obs.AwsAuthentication{
						Type: obs.AwsAuthTypeIAMRole,
						IamRole: &obs.AwsRole{
							RoleARN: obs.SecretReference{
								Key:        constants.AwsWebIdentityRoleKey,
								SecretName: secretNameWebIdentity,
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obs.AwsAssumeRole{
							RoleARN: obs.SecretReference{
								Key:        "assume_role_arn",
								SecretName: secretNameWebIdentity,
							},
							ExternalID: "unique-web-external-id",
						},
					},
				}
			}, framework.NoOptions, "files/s3_with_aws_credentials.toml"),

			Entry("when URL is spec'd", func(spec *obs.OutputSpec) {
				spec.S3.KeyPrefix = "app-{.log_type||\"missing\"}"
				spec.S3.URL = "http://mylogreceiver"
			}, framework.NoOptions, "files/s3_with_url.toml"),
		)
	})
})
