package observability_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/test"
	"strings"
)

var _ = Describe("helpers for output types", func() {

	Context("#SecretReferences", func() {

		It("should return an empty set of keys when authentication is not defined for an output", func() {
			for _, t := range obsv1.OutputTypes {

				outputType := strings.TrimPrefix("OutputType", string(t))
				outputType = strings.ToLower(outputType[0:1]) + outputType[1:]
				yaml := test.JSONLine(map[string]interface{}{
					"type":     t,
					outputType: map[string]interface{}{},
				})
				spec := &obsv1.OutputSpec{}
				test.MustUnmarshal(yaml, spec)
				Expect(SecretReferences(*spec)).To(BeEmpty())
			}
		})

	})
})

var _ = Describe("AzureLogsIngestion secret handling", func() {
	Context("SecretReferences", func() {
		It("should return client secret reference for client secret authentication", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeAzureLogsIngestion,
				AzureLogsIngestion: &obsv1.AzureLogsIngestion{
					Authentication: &obsv1.AzureLogsIngestionAuthentication{
						Type: obsv1.AzureLogsIngestionAuthTypeClientSecret,
						ClientSecret: &obsv1.AzureLogsIngestionClientSecret{
							TenantId: "tenant-id",
							ClientId: "client-id",
							Secret: &obsv1.SecretReference{
								SecretName: "azure-secret",
								Key:        "client_secret",
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(1))
			Expect(refs[0].SecretName).To(Equal("azure-secret"))
			Expect(refs[0].Key).To(Equal("client_secret"))
		})

		It("should return token secret reference for workload identity with secret-sourced token", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeAzureLogsIngestion,
				AzureLogsIngestion: &obsv1.AzureLogsIngestion{
					Authentication: &obsv1.AzureLogsIngestionAuthentication{
						Type: obsv1.AzureLogsIngestionAuthTypeWorkloadIdentity,
						WorkloadIdentity: &obsv1.AzureLogsIngestionWorkloadIdentity{
							TenantId: "tenant-id",
							ClientId: "client-id",
							Token: &obsv1.BearerToken{
								From: obsv1.BearerTokenFromSecret,
								Secret: &obsv1.BearerTokenSecretKey{
									Name: "token-secret",
									Key:  "token",
								},
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(1))
			Expect(refs[0].SecretName).To(Equal("token-secret"))
			Expect(refs[0].Key).To(Equal("token"))
		})

		It("should return empty for workload identity with service account token", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeAzureLogsIngestion,
				AzureLogsIngestion: &obsv1.AzureLogsIngestion{
					Authentication: &obsv1.AzureLogsIngestionAuthentication{
						Type: obsv1.AzureLogsIngestionAuthTypeWorkloadIdentity,
						WorkloadIdentity: &obsv1.AzureLogsIngestionWorkloadIdentity{
							TenantId: "tenant-id",
							ClientId: "client-id",
							Token: &obsv1.BearerToken{
								From: obsv1.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(BeEmpty())
		})
	})

	Context("NeedServiceAccountToken", func() {
		It("should return true for workload identity with service account token", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeAzureLogsIngestion,
					AzureLogsIngestion: &obsv1.AzureLogsIngestion{
						Authentication: &obsv1.AzureLogsIngestionAuthentication{
							Type: obsv1.AzureLogsIngestionAuthTypeWorkloadIdentity,
							WorkloadIdentity: &obsv1.AzureLogsIngestionWorkloadIdentity{
								TenantId: "tenant-id",
								ClientId: "client-id",
								Token: &obsv1.BearerToken{
									From: obsv1.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeTrue())
		})

		It("should return false for client secret authentication", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeAzureLogsIngestion,
					AzureLogsIngestion: &obsv1.AzureLogsIngestion{
						Authentication: &obsv1.AzureLogsIngestionAuthentication{
							Type: obsv1.AzureLogsIngestionAuthTypeClientSecret,
							ClientSecret: &obsv1.AzureLogsIngestionClientSecret{
								TenantId: "tenant-id",
								ClientId: "client-id",
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeFalse())
		})
	})
})

var _ = Describe("S3 secret handling", func() {
	Context("SecretReferences", func() {
		It("should return access key secrets for S3 with access key authentication", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeS3,
				S3: &obsv1.S3{
					Authentication: &obsv1.AwsAuthentication{
						Type: obsv1.AwsAuthTypeAccessKey,
						AwsAccessKey: &obsv1.AwsAccessKey{
							KeyId: obsv1.SecretReference{
								SecretName: "s3-secret1",
								Key:        "access-key-id",
							},
							KeySecret: obsv1.SecretReference{
								SecretName: "s3-secret2",
								Key:        "secret-access-key",
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(2))
			Expect(refs[0].SecretName).To(Equal("s3-secret1"))
			Expect(refs[0].Key).To(Equal("access-key-id"))
			Expect(refs[1].SecretName).To(Equal("s3-secret2"))
			Expect(refs[1].Key).To(Equal("secret-access-key"))
		})

		It("should return IAM role secrets for S3 with IAM role authentication", func() {
			output := obsv1.OutputSpec{
				Type: obsv1.OutputTypeS3,
				S3: &obsv1.S3{
					Authentication: &obsv1.AwsAuthentication{
						Type: obsv1.AwsAuthTypeIAMRole,
						IamRole: &obsv1.AwsRole{
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
					Authentication: &obsv1.AwsAuthentication{
						Type: obsv1.AwsAuthTypeIAMRole,
						IamRole: &obsv1.AwsRole{
							RoleARN: obsv1.SecretReference{
								SecretName: "s3-base-role",
								Key:        "role-arn",
							},
							Token: obsv1.BearerToken{
								From: obsv1.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obsv1.AwsAssumeRole{
							RoleARN: obsv1.SecretReference{
								SecretName: "s3-assume-role",
								Key:        "role-arn",
							},
						},
					},
				},
			}

			refs := SecretReferences(output)
			Expect(refs).To(HaveLen(2))
			Expect(refs[0].SecretName).To(Equal("s3-base-role"))
			Expect(refs[0].Key).To(Equal("role-arn"))
			Expect(refs[1].SecretName).To(Equal("s3-assume-role"))
			Expect(refs[1].Key).To(Equal("role-arn"))
		})
	})

	Context("NeedServiceAccountToken", func() {
		It("should return true for S3 with IAM role authentication using service account token", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeS3,
					S3: &obsv1.S3{
						Authentication: &obsv1.AwsAuthentication{
							Type: obsv1.AwsAuthTypeIAMRole,
							IamRole: &obsv1.AwsRole{
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
						Authentication: &obsv1.AwsAuthentication{
							Type: obsv1.AwsAuthTypeAccessKey,
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeFalse())
		})

		It("should return true for GCP output type regardless of non-sts key", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeGoogleCloudLogging,
					GoogleCloudLogging: &obsv1.GoogleCloudLogging{
						Authentication: &obsv1.GoogleCloudLoggingAuthentication{
							Credentials: &obsv1.SecretReference{
								SecretName: "gcp-creds",
								Key:        "service_account.json",
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeTrue())
		})

		It("should return true for GCP output type regardless of key", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeGoogleCloudLogging,
					GoogleCloudLogging: &obsv1.GoogleCloudLogging{
						Authentication: &obsv1.GoogleCloudLoggingAuthentication{
							Credentials: &obsv1.SecretReference{
								SecretName: "gcp-creds",
								Key:        "external_account.json",
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeTrue())
		})

		It("should return true for Loki with service account token", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeLoki,
					Loki: &obsv1.Loki{
						Authentication: &obsv1.HTTPAuthentication{
							Token: &obsv1.BearerToken{
								From: obsv1.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeTrue())
		})

		It("should return false for Loki with token from secret", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeLoki,
					Loki: &obsv1.Loki{
						Authentication: &obsv1.HTTPAuthentication{
							Token: &obsv1.BearerToken{
								From: obsv1.BearerTokenFromSecret,
								Secret: &obsv1.BearerTokenSecretKey{
									Name: "token-secret",
									Key:  "token",
								},
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeFalse())
		})

		It("should return true when at least one output needs service account token", func() {
			outputs := Outputs{
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeS3,
					S3: &obsv1.S3{
						Authentication: &obsv1.AwsAuthentication{
							Type: obsv1.AwsAuthTypeAccessKey,
						},
					},
				},
				obsv1.OutputSpec{
					Type: obsv1.OutputTypeLoki,
					Loki: &obsv1.Loki{
						Authentication: &obsv1.HTTPAuthentication{
							Token: &obsv1.BearerToken{
								From: obsv1.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			Expect(outputs.NeedServiceAccountToken()).To(BeTrue())
		})
	})
})
