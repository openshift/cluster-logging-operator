package cloudwatch_test

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("cloudwatch auth configmap", func() {
	const (
		roleArn     = "arn:aws:iam::123456789012:role/foo"
		roleArn2    = "arn:aws:iam::123456789012:role/bar"
		saTokenPath = "/var/run/ocp-collector/serviceaccount/token"
		region      = "us-west-1"
	)
	Context("generating ProfileCredentials objects", func() {
		var (
			cwSecret map[string]*v1.Secret
		)

		BeforeEach(func() {
			cwSecret = map[string]*v1.Secret{
				"cw-secret": {
					Data: map[string][]byte{
						"role_arn1": []byte(roleArn),
						"role_arn2": []byte(roleArn2),
						"token":     []byte("my-token"),
					},
				},
			}
		})

		It("should be nil if no cloudwatch outputs", func() {
			outputs := []obs.OutputSpec{
				{
					Name: "es-out",
					Type: obs.OutputTypeElasticsearch,
				},
			}
			Expect(cloudwatch.GenerateProfileCreds(nil, "test-clf", outputs, cwSecret)).To(BeNil())
		})

		It("should be nil if secrets are nil and no cloudwatch outputs", func() {
			outputs := []obs.OutputSpec{
				{
					Name: "es-out",
					Type: obs.OutputTypeElasticsearch,
				},
			}

			Expect(cloudwatch.GenerateProfileCreds(nil, "test-clf", outputs, nil)).To(BeNil())
		})

		It("should be nil if cloudwatch output is not role based", func() {
			outputs := []obs.OutputSpec{
				{
					Name: "my-cw",
					Type: obs.OutputTypeCloudwatch,
					Cloudwatch: &obs.Cloudwatch{
						Authentication: &obs.CloudwatchAuthentication{
							Type: obs.CloudwatchAuthTypeAccessKey,
							AWSAccessKey: &obs.CloudwatchAWSAccessKey{
								KeySecret: obs.SecretReference{
									Key:        "role_arn1",
									SecretName: "cw-secret",
								},
							},
						},
					},
				},
			}

			Expect(cloudwatch.GenerateProfileCreds(nil, "test-clf", outputs, nil)).To(BeNil())
		})

		It("should be nil if secrets are nil and outputs are nil", func() {
			Expect(cloudwatch.GenerateProfileCreds(nil, "test-clf", nil, nil)).To(BeNil())
		})

		DescribeTable("token path", func(token obs.BearerToken, exp cloudwatch.ProfileCredentials) {
			cwOutputs := []obs.OutputSpec{
				{
					Name: "cw-out",
					Type: obs.OutputTypeCloudwatch,
					Cloudwatch: &obs.Cloudwatch{
						Authentication: &obs.CloudwatchAuthentication{
							Type: obs.CloudwatchAuthTypeIAMRole,
							IAMRole: &obs.CloudwatchIAMRole{
								RoleARN: obs.SecretReference{
									Key:        "role_arn1",
									SecretName: "cw-secret",
								},
								Token: token,
							},
						},
						Region: region,
					},
				},
			}
			actIds := cloudwatch.GenerateProfileCreds(nil, "test-clf", cwOutputs, cwSecret)
			Expect(actIds[0]).To(Equal(exp))
		},
			Entry("should get token from secret", obs.BearerToken{
				From: obs.BearerTokenFromSecret,
				Secret: &obs.BearerTokenSecretKey{
					Key:  constants.TokenKey,
					Name: "cw-secret",
				},
			}, cloudwatch.ProfileCredentials{
				Name:                 "cw-out",
				RoleARN:              roleArn,
				WebIdentityTokenFile: "/var/run/ocp-collector/secrets/cw-secret/token",
			}),
			Entry("should get token from serviceAccount", obs.BearerToken{
				From: obs.BearerTokenFromServiceAccount,
			}, cloudwatch.ProfileCredentials{
				Name:                 "cw-out",
				RoleARN:              roleArn,
				WebIdentityTokenFile: "/var/run/ocp-collector/serviceaccount/token",
			}))

		It("should gather all role_arns/tokens from cw outputs", func() {
			cwOutputs := []obs.OutputSpec{
				{
					Name: "cw-out1",
					Type: obs.OutputTypeCloudwatch,
					Cloudwatch: &obs.Cloudwatch{
						Authentication: &obs.CloudwatchAuthentication{
							Type: obs.CloudwatchAuthTypeIAMRole,
							IAMRole: &obs.CloudwatchIAMRole{
								RoleARN: obs.SecretReference{
									Key:        "role_arn1",
									SecretName: "cw-secret",
								},
								Token: obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
						Region: region,
					},
				},
				{
					Name: "cw-out2",
					Type: obs.OutputTypeCloudwatch,
					Cloudwatch: &obs.Cloudwatch{
						Authentication: &obs.CloudwatchAuthentication{
							Type: obs.CloudwatchAuthTypeIAMRole,
							IAMRole: &obs.CloudwatchIAMRole{
								RoleARN: obs.SecretReference{
									Key:        "role_arn2",
									SecretName: "cw-secret",
								},
								Token: obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
						Region: region,
					},
				},
			}

			expCreds := []cloudwatch.ProfileCredentials{
				{
					Name:                 cwOutputs[0].Name,
					RoleARN:              roleArn,
					WebIdentityTokenFile: saTokenPath,
				},
				{
					Name:                 cwOutputs[1].Name,
					RoleARN:              roleArn2,
					WebIdentityTokenFile: saTokenPath,
				},
			}

			actIds := cloudwatch.GenerateProfileCreds(nil, "test-clf", cwOutputs, cwSecret)
			Expect(actIds).To(Equal(expCreds))
		})
	})

	DescribeTable("cloudwatch credential go template", func(creds []cloudwatch.ProfileCredentials, expFile string) {
		exp, err := testFiles.ReadFile(expFile)
		Expect(err).To(BeNil())

		w := &strings.Builder{}
		err = cloudwatch.ProfileTemplate.Execute(w, creds)

		Expect(err).To(BeNil())
		Expect(w.String()).To(Equal(string(exp)))
	},
		Entry("should generate one profile", []cloudwatch.ProfileCredentials{
			{
				Name:                 "default",
				RoleARN:              "arn:aws:iam::123456789012:role/test-default",
				WebIdentityTokenFile: saTokenPath,
			},
		}, "files/cw_single_credential"),
		Entry("should generate multiple profiles when multiple credentials are present", []cloudwatch.ProfileCredentials{
			{
				Name:                 "default",
				RoleARN:              "arn:aws:iam::123456789012:role/test-default",
				WebIdentityTokenFile: saTokenPath,
			},
			{
				Name:                 "foo",
				RoleARN:              "arn:aws:iam::123456789012:role/test-foo",
				WebIdentityTokenFile: saTokenPath,
			},
			{
				Name:                 "bar",
				RoleARN:              "arn:aws:iam::123456789012:role/test-bar",
				WebIdentityTokenFile: saTokenPath,
			},
		}, "files/cw_multiple_credentials"),
		Entry("should generate assume role profile", []cloudwatch.ProfileCredentials{
			{
				Name:                 "default",
				RoleARN:              "arn:aws:iam::123456789012:role/test-default",
				WebIdentityTokenFile: saTokenPath,
				AssumeRole:           "arn:aws:iam::987654321098:role/cross-account-role",
				ExternalID:           "unique-external-id",
				SessionName:          "output-default",
			},
		}, "files/cw_assume_role_single"))
})
