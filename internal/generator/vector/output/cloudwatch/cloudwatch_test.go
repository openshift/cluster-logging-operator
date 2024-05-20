package cloudwatch_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating vector config for cloudwatch output", func() {

	const (
		roleArn    = "arn:aws:iam::123456789012:role/my-role-to-assume"
		secretName = "vector-cw-secret"
	)
	Context("#New", func() {

		const (
			groupPrefix           = "all-logs"
			keyId                 = "AKIAIOSFODNN7EXAMPLE"
			keySecret             = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
			secretName            = "vector-cw-secret"
			secretWithCredentials = "secretwithcredentials"
			vectorTLSSecret       = "vector-cw-secret-tls"
		)

		var (
			tlsSpec = &obs.OutputTLSSpec{
				InsecureSkipVerify: false,
				TLSSpec: obs.TLSSpec{
					CA: &obs.ConfigMapOrSecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: vectorTLSSecret,
						},
						Key: constants.TrustedCABundleKey,
					},
					Certificate: &obs.ConfigMapOrSecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: vectorTLSSecret,
						},
						Key: constants.ClientCertKey,
					},
					Key: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: vectorTLSSecret,
						},
						Key: constants.ClientPrivateKey,
					},
				},
			}

			initOutput = func() obs.OutputSpec {
				return obs.OutputSpec{
					Type: loggingv1.OutputTypeCloudwatch,
					Name: "cw",
					Cloudwatch: &obs.Cloudwatch{
						Region: "us-east-test",
						Authentication: &obs.CloudwatchAuthentication{
							AccessKeyID: &obs.SecretKey{
								Secret: &corev1.LocalObjectReference{
									Name: secretName,
								},
								Key: constants.AWSAccessKeyID,
							},
							AccessKeySecret: &obs.SecretKey{
								Secret: &corev1.LocalObjectReference{
									Name: secretName,
								},
								Key: constants.AWSSecretAccessKey,
							},
						},
					},
				}
			}

			secrets = map[string]*corev1.Secret{
				secretName: {
					Data: map[string][]byte{
						constants.AWSAccessKeyID:     []byte(keyId),
						constants.AWSSecretAccessKey: []byte(keySecret),
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
						constants.AWSCredentialsKey:  []byte("[default]\nrole_arn = " + roleArn + "\nweb_identity_token_file = /var/run/secrets/token"),
						constants.ClientPrivateKey:   []byte("-- key-- "),
						constants.TrustedCABundleKey: []byte("-- ca-bundle -- "),
					},
				},
			}
		)

		DescribeTable("should generate valid config", func(groupPrefix string, groupByType obs.LogGroupByType, visit func(spec *obs.OutputSpec), op framework.Options, expFile string) {
			exp, err := tomlContent.ReadFile(expFile)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
			}
			outputSpec := initOutput()
			outputSpec.Cloudwatch.GroupPrefix = groupPrefix
			outputSpec.Cloudwatch.GroupBy = groupByType
			if visit != nil {
				visit(&outputSpec)
			}
			conf := New(outputSpec.Name, outputSpec, []string{"cw-forward"}, secrets, nil, op)
			Expect(string(exp)).To(EqualConfigFrom(conf))
		},
			Entry("when prefix and group by log_type is spec'd", groupPrefix, obs.LogGroupByLogType, nil, framework.NoOptions, "cw_group_prefix_by_log_type.toml"),
			Entry("when prefix and group by namespace name is spec'd", groupPrefix, obs.LogGroupByNamespaceName, nil, framework.NoOptions, "cw_group_prefix_by_namespace_name.toml"),
			Entry("when prefix and group by namespace uuid is spec'd", groupPrefix, obs.LogGroupByNamespaceUUID, nil, framework.NoOptions, "cw_group_prefix_by_namespace_uuid.toml"),
			Entry("when prefix is nil and group by log_type is spec'd", "", obs.LogGroupByLogType, nil, framework.NoOptions, "cw_no_group_prefix_by_log_type.toml"),
			Entry("when URL is spec'd", "", obs.LogGroupByLogType, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.URL = "http://mylogreceiver"
			}, framework.NoOptions, "cw_with_url.toml"),
			Entry("when minTLS and ciphers is spec'd", "", obs.LogGroupByLogType, nil, testhelpers.FrameworkOptionWithDefaultTLSCiphers, "cw_with_tls_and_default_mintls_ciphers.toml"),
			Entry("when tls is spec'd", "", obs.LogGroupByLogType, func(spec *obs.OutputSpec) {
				spec.TLS = tlsSpec
			}, framework.NoOptions, "cw_with_tls_spec.toml"),
			Entry("when tls is spec'd with insecure verify", "", obs.LogGroupByLogType, func(spec *obs.OutputSpec) {
				spec.TLS = tlsSpec
				spec.TLS.InsecureSkipVerify = true
			}, framework.NoOptions, "cw_with_tls_spec_insecure_verify.toml"),
			Entry("when aws credentials are provided", groupPrefix, obs.LogGroupByLogType, func(spec *obs.OutputSpec) {
				spec.Cloudwatch.Authentication = &obs.CloudwatchAuthentication{
					Credentials: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: secretWithCredentials,
						},
						Key: constants.AWSCredentialsKey,
					},
				}
			}, framework.NoOptions, "cw_group_prefix_with_aws_credentials.toml"),
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
						constants.AWSWebIdentityRoleKey: []byte(roleArn),
						"altArn":                        []byte(altRoleArn),
						constants.AWSCredentialsKey:     []byte(credentialsString),
						"badArn":                        []byte("no match here"),
						"role_arn_as_cred":              []byte(credentialsString),
					},
				},
			}
		)
		DescribeTable("when retrieving the role_arn", func(auth obs.CloudwatchAuthentication, exp string) {
			results := ParseRoleArn(&auth, secrets)
			Expect(results).To(Equal(exp))
		},
			Entry("should return the value explicity spec'd",
				obs.CloudwatchAuthentication{
					RoleARN: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: constants.AWSWebIdentityRoleKey,
					},
				}, roleArn),
			Entry("should return a specified valid role_arn when the partition is more than 'aws'",
				obs.CloudwatchAuthentication{
					RoleARN: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: "altArn",
					},
				}, altRoleArn),
			Entry("should return a valid role_arn when using 'credentials' ",
				obs.CloudwatchAuthentication{
					Credentials: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: constants.AWSCredentialsKey,
					},
				}, credentialsRoleArn),
			Entry("should return the value from the credentials string when specified as role_arn",
				obs.CloudwatchAuthentication{
					RoleARN: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: "role_arn_as_cred",
					},
				}, credentialsRoleArn),
			Entry("should return an empty string when value is incorrectly formatted",
				obs.CloudwatchAuthentication{
					Credentials: &obs.SecretKey{
						Secret: &corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: "bad",
					},
				}, ""),
		)

	})

})
