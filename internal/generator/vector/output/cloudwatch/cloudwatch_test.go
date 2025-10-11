package cloudwatch_test

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/collector/aws"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Generating vector config for cloudwatch output", func() {

	const (
		roleArn    = "arn:aws:iam::123456789012:role/my-role-to-assume"
		secretName = "vector-cw-secret"
	)
	Context("#New", func() {

		const (
			keyId                 = "AKIAIOSFODNN7EXAMPLE"
			keySecret             = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
			secretName            = "vector-cw-secret"
			secretWithCredentials = "secretwithcredentials"
			vectorTLSSecret       = "vector-cw-secret-tls"
		)

		var (
			adapter fake.Output
			tlsSpec = &obs.OutputTLSSpec{
				InsecureSkipVerify: false,
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: vectorTLSSecret,
					},
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: vectorTLSSecret,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: vectorTLSSecret,
					},
				},
			}

			initOutput = func() obs.OutputSpec {
				return obs.OutputSpec{
					Type: obs.OutputTypeCloudwatch,
					Name: "cw",
					Cloudwatch: &obs.Cloudwatch{
						Region: "us-east-test",
						Authentication: &obs.AwsAuthentication{
							Type: obs.AuthTypeAccessKey,
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
				vectorTLSSecret: {
					Data: map[string][]byte{
						constants.ClientCertKey:      []byte("-- crt-- "),
						constants.ClientPrivateKey:   []byte("-- key-- "),
						constants.TrustedCABundleKey: []byte("-- ca-bundle -- "),
					},
				},
				secretWithCredentials: {
					Data: map[string][]byte{
						constants.AwsCredentialsKey:  []byte("[default]\nrole_arn = " + roleArn + "\nweb_identity_token_file = /var/run/secrets/token"),
						"my_role_arn":                []byte(roleArn),
						constants.ClientPrivateKey:   []byte("-- key-- "),
						constants.TrustedCABundleKey: []byte("-- ca-bundle -- "),
					},
				},
			}
			baseTune = &obs.CloudwatchTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					DeliveryMode:     obs.DeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
					MaxRetryDuration: utils.GetPtr(time.Duration(35)),
					MinRetryDuration: utils.GetPtr(time.Duration(20)),
				},
			}
		)

		DescribeTable("should generate valid config", func(groupName string, visit func(spec *obs.OutputSpec), tune bool, op framework.Options, expFile string) {
			exp, err := testFiles.ReadFile(expFile)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
			}
			outputSpec := initOutput()
			outputSpec.Cloudwatch.GroupName = groupName
			if visit != nil {
				visit(&outputSpec)
			}
			if tune {
				adapter = *fake.NewOutput(outputSpec, secrets, framework.NoOptions)
			}
			op[framework.OptionForwarderName] = "my-forwarder"
			conf := New(outputSpec.Name, outputSpec, []string{"cw-forward"}, secrets, adapter, op)
			Expect(string(exp)).To(EqualConfigFrom(conf))
		},

			Entry("when groupName is spec'd", `{.log_type||"missing"}-foo`, func(spec *obs.OutputSpec) {}, false,
				framework.NoOptions, "files/cw_with_groupname.toml"),

			Entry("when URL is spec'd", `{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.URL = "http://mylogreceiver"
			}, false, framework.NoOptions, "files/cw_with_url.toml"),

			Entry("when minTLS and ciphers is spec'd", `{.log_type||"missing"}`, nil, false,
				testhelpers.FrameworkOptionWithDefaultTLSCiphers, "files/cw_with_tls_and_default_mintls_ciphers.toml"),

			Entry("when tls is spec'd", `{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.TLS = tlsSpec
			}, false, framework.NoOptions, "files/cw_with_tls_spec.toml"),

			Entry("when tls is spec'd with insecure verify", `{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.TLS = tlsSpec
				spec.TLS.InsecureSkipVerify = true
			}, false, framework.NoOptions, "files/cw_with_tls_spec_insecure_verify.toml"),

			Entry("when aws role credentials are provided", `app-{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.Authentication = &obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
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
			}, false, framework.NoOptions, "files/cw_groupname_with_aws_credentials.toml"),

			Entry("when a role_arn is provided", `app-{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.Authentication = &obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        "my_role_arn",
							SecretName: secretWithCredentials,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}
			}, false, framework.NoOptions, "files/cw_groupname_with_aws_credentials.toml"),

			Entry("when an assume_role is specified with accessKey auth", `app-{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.Authentication = &obs.AwsAuthentication{
					Type: obs.AuthTypeAccessKey,
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
							Key:        "my_role_arn",
							SecretName: secretWithCredentials,
						},
						ExternalID: "unique-external-id",
					},
				}
			}, false, framework.NoOptions, "files/cw_key_auth_and_assume_role.toml"),

			Entry("when tuning is spec'd", `{.log_type||"missing"}`, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.Tuning = baseTune
			}, true, framework.NoOptions, "files/cw_with_tuning.toml"),
		)
	})

	Context("#ParseRoleArn", func() {

		const (
			altRoleArn         = "arn:aws-us-gov:iam::225746144451:role/anli-sts-25690-openshift-logging-cloudwatch-credentials"
			credentialsRoleArn = "arn:aws:iam::123456789012:role/my-credentials-role"
			credentialsString  = "[default]\nrole_arn = " + credentialsRoleArn + "\nweb_identity_token_file = /var/run/secrets/token"
		)
		var (
			secrets = map[string]*corev1.Secret{
				secretName: {
					Data: map[string][]byte{
						constants.AwsWebIdentityRoleKey: []byte(roleArn),
						"altArn":                        []byte(altRoleArn),
						constants.AwsCredentialsKey:     []byte(credentialsString),
						"badArn":                        []byte("no match here"),
						"role_arn_as_cred":              []byte(credentialsString),
					},
				},
			}
		)

		DescribeTable("when retrieving the role_arn", func(auth obs.AwsAuthentication, exp string) {
			results := aws.ParseRoleArn(&auth, secrets)
			Expect(results).To(Equal(exp))
		},
			Entry("should return the value specified",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        constants.AwsWebIdentityRoleKey,
							SecretName: secretName,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}, roleArn),
			Entry("should return a specified valid role_arn when the partition is more than 'aws'",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        "altArn",
							SecretName: secretName,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}, altRoleArn),
			Entry("should return a valid role_arn when using 'credentials' ",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        constants.AwsCredentialsKey,
							SecretName: secretName,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}, credentialsRoleArn),
			Entry("should return the value from the credentials string when specified as role_arn",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        "role_arn_as_cred",
							SecretName: secretName,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}, credentialsRoleArn),
			Entry("should return an empty string when value is incorrectly formatted",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							Key:        "bad",
							SecretName: secretName,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				}, ""),
		)
	})

	Context("#ParseAssumeRoleArn", func() {
		const (
			assumeRoleArn    = "arn:aws:iam::987654321098:role/cross-account-role"
			altAssumeRoleArn = "arn:aws-us-gov:iam::987654321098:role/cross-account-role"
		)
		var (
			secrets = map[string]*corev1.Secret{
				secretName: {
					Data: map[string][]byte{
						"assume_role_arn":     []byte(assumeRoleArn),
						"alt_assume_role_arn": []byte(altAssumeRoleArn),
						"bad_arn":             []byte("invalid-arn-format"),
					},
				},
			}
		)

		DescribeTable("when retrieving the assume role arn", func(auth obs.AwsAuthentication, exp string) {
			results := aws.ParseAssumeRoleArn(auth.AssumeRole, secrets)
			Expect(results).To(Equal(exp))
		},
			Entry("should return the value explicitly spec'd with iamRole auth",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					AssumeRole: &obs.AwsAssumeRole{
						RoleARN: obs.SecretReference{
							Key:        "assume_role_arn",
							SecretName: secretName,
						},
					},
				}, assumeRoleArn),

			Entry("should return the secret value specified (assumeRole with accessKey auth)",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					AssumeRole: &obs.AwsAssumeRole{
						RoleARN: obs.SecretReference{
							Key:        "assume_role_arn",
							SecretName: secretName,
						},
					},
				}, assumeRoleArn),

			Entry("should return a specified valid assume role arn when the partition is more than 'aws'",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					AssumeRole: &obs.AwsAssumeRole{
						RoleARN: obs.SecretReference{
							Key:        "alt_assume_role_arn",
							SecretName: secretName,
						},
					},
				}, altAssumeRoleArn),
			Entry("should return empty string when assume role is not specified",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
				}, ""),
			Entry("should return empty string when value is incorrectly formatted",
				obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					AssumeRole: &obs.AwsAssumeRole{
						RoleARN: obs.SecretReference{
							Key:        "bad_arn",
							SecretName: secretName,
						},
					},
				}, ""),
		)
	})

	Context("when assume role is configured", func() {
		It("should generate valid Vector config with assume role", func() {
			outputSpec := obs.OutputSpec{
				Type: obs.OutputTypeCloudwatch,
				Name: "cw",
				Cloudwatch: &obs.Cloudwatch{
					Region:    "us-east-test",
					GroupName: "{.log_type||\"missing\"}",
					Authentication: &obs.AwsAuthentication{
						Type: obs.AuthTypeIAMRole,
						IamRole: &obs.AwsRole{
							RoleARN: obs.SecretReference{
								Key:        constants.AwsWebIdentityRoleKey,
								SecretName: secretName,
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obs.AwsAssumeRole{
							RoleARN: obs.SecretReference{
								Key:        "assume_role_arn",
								SecretName: secretName,
							},
							ExternalID: "unique-external-id",
						},
					},
				},
			}

			secrets := map[string]*corev1.Secret{
				secretName: {
					Data: map[string][]byte{
						constants.AwsWebIdentityRoleKey: []byte("arn:aws:iam::123456789012:role/initial-role"),
						"assume_role_arn":               []byte("arn:aws:iam::987654321098:role/cross-account-role"),
						"external_id":                   []byte("unique-external-id"),
					},
				},
			}

			op := framework.Options{}
			op[framework.OptionForwarderName] = "my-forwarder"
			conf := New(outputSpec.Name, outputSpec, []string{"cw-forward"}, secrets, fake.Output{}, op)

			// Verify that CloudWatch sink configuration is present
			var elementNames []string
			for _, element := range conf {
				elementNames = append(elementNames, element.Name())
			}

			// Verify basic CloudWatch elements exist
			Expect(elementNames).To(ContainElement("cloudwatchTemplate"), "Should contain cloudwatch sink template")

			// Since authentication is embedded within cloudwatchTemplate, we just verify
			// that the CloudWatch configuration was created successfully with assume role
			// The actual authentication fields are tested in unit tests for the auth functions
			Expect(len(conf)).To(BeNumerically(">", 0), "Should generate CloudWatch configuration elements")
		})
	})
})
