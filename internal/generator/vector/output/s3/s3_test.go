package s3

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("S3 output", func() {
	var (
		secrets map[string]*corev1.Secret
	)

	BeforeEach(func() {
		secrets = map[string]*corev1.Secret{
			"s3-access-key": {
				Data: map[string][]byte{
					"aws_access_key_id":     []byte("access_key_id"),
					"aws_secret_access_key": []byte("secret_access_key"),
				},
			},
			"s3-role": {
				Data: map[string][]byte{
					"role-arn": []byte("arn:aws:iam::123456789012:role/my-role"),
				},
			},
			"s3-assume-role": {
				Data: map[string][]byte{
					"role-arn":    []byte("arn:aws:iam::123456789012:role/assume-role"),
					"external-id": []byte("external-123"),
				},
			},
		}
	})

	Context("with access key authentication", func() {
		It("should generate valid configuration", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Region: "us-east-1",
					Bucket: "my-log-bucket",
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

			elements := New("test", output, []string{"input"}, secrets, nil, framework.NoOptions)
			Expect(elements).To(HaveLen(8))

			// Check that we have the expected elements
			sink := elements[1].(*S3)
			Expect(sink.ComponentID).To(Equal("test"))
			Expect(sink.Region).To(Equal("us-east-1"))
			Expect(sink.Bucket).To(Equal("my-log-bucket"))
		})
	})

	Context("with IAM role authentication", func() {
		It("should generate valid configuration with service account token", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Region: "us-east-1",
					Bucket: "my-log-bucket",
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "s3-role",
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			elements := New("test", output, []string{"input"}, secrets, nil, framework.Options{"forwarderName": "test-forwarder"})
			Expect(elements).To(HaveLen(8))

			sink := elements[1].(*S3)
			Expect(sink.ComponentID).To(Equal("test"))
			Expect(sink.Region).To(Equal("us-east-1"))
			Expect(sink.Bucket).To(Equal("my-log-bucket"))
		})

		It("should generate valid configuration with assume role", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Region: "us-east-1",
					Bucket: "my-log-bucket",
					Authentication: &obs.S3Authentication{
						Type: obs.S3AuthTypeIAMRole,
						IAMRole: &obs.S3IAMRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "s3-role",
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
						AssumeRole: &obs.CloudwatchAssumeRole{
							RoleARN: obs.SecretReference{
								Key:        "role-arn",
								SecretName: "s3-assume-role",
							},
							ExternalID: &obs.SecretReference{
								Key:        "external-id",
								SecretName: "s3-assume-role",
							},
							SessionName: "s3-session",
						},
					},
				},
			}

			elements := New("test", output, []string{"input"}, secrets, nil, framework.Options{"forwarderName": "test-forwarder"})
			Expect(elements).To(HaveLen(8))

			sink := elements[1].(*S3)
			Expect(sink.ComponentID).To(Equal("test"))
			Expect(sink.Region).To(Equal("us-east-1"))
			Expect(sink.Bucket).To(Equal("my-log-bucket"))
		})
	})

	Context("with key prefix", func() {
		It("should generate valid configuration with static key prefix", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Region:    "us-east-1",
					Bucket:    "my-log-bucket",
					KeyPrefix: "logs/",
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

			elements := New("test", output, []string{"input"}, secrets, nil, framework.NoOptions)
			Expect(elements).To(HaveLen(8))

			sink := elements[1].(*S3)
			Expect(sink.KeyPrefix).To(Equal("test_key_prefix"))
		})

		It("should generate valid configuration with timestamp-based key prefix", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Region:    "us-east-1",
					Bucket:    "my-log-bucket",
					KeyPrefix: "logs/{@timestamp|year}/{@timestamp|month}/{@timestamp|day}/",
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

			elements := New("test", output, []string{"input"}, secrets, nil, framework.NoOptions)
			Expect(elements).To(HaveLen(8))

			sink := elements[1].(*S3)
			Expect(sink.KeyPrefix).To(Equal("test_key_prefix"))
		})
	})

	Context("with custom endpoint", func() {
		It("should generate valid configuration with custom S3-compatible endpoint", func() {
			output := obs.OutputSpec{
				Type: obs.OutputTypeS3,
				Name: "s3-output",
				S3: &obs.S3{
					Region: "us-east-1",
					Bucket: "my-log-bucket",
					URL:    "https://s3.custom-endpoint.com",
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

			elements := New("test", output, []string{"input"}, secrets, nil, framework.NoOptions)
			Expect(elements).To(HaveLen(8))

			sink := elements[1].(*S3)
			Expect(sink.ComponentID).To(Equal("test"))
			Expect(sink.Region).To(Equal("us-east-1"))
			Expect(sink.Bucket).To(Equal("my-log-bucket"))

			// Check endpoint configuration
			endpoint := sink.EndpointConfig.(Endpoint)
			Expect(endpoint.URL).To(Equal("https://s3.custom-endpoint.com"))
		})
	})

	Context("ARN parsing", func() {
		It("should parse valid role ARN", func() {
			auth := &obs.S3Authentication{
				Type: obs.S3AuthTypeIAMRole,
				IAMRole: &obs.S3IAMRole{
					RoleARN: obs.SecretReference{
						Key:        "role-arn",
						SecretName: "s3-role",
					},
				},
			}

			arn := ParseRoleArn(auth, secrets)
			Expect(arn).To(Equal("arn:aws:iam::123456789012:role/my-role"))
		})

		It("should parse valid assume role ARN", func() {
			auth := &obs.S3Authentication{
				AssumeRole: &obs.CloudwatchAssumeRole{
					RoleARN: obs.SecretReference{
						Key:        "role-arn",
						SecretName: "s3-assume-role",
					},
				},
			}

			arn := ParseAssumeRoleArn(auth, secrets)
			Expect(arn).To(Equal("arn:aws:iam::123456789012:role/assume-role"))
		})

		It("should return empty string for invalid ARN", func() {
			auth := &obs.S3Authentication{
				Type: obs.S3AuthTypeIAMRole,
				IAMRole: &obs.S3IAMRole{
					RoleARN: obs.SecretReference{
						Key:        "invalid-key",
						SecretName: "s3-role",
					},
				},
			}

			arn := ParseRoleArn(auth, secrets)
			Expect(arn).To(BeEmpty())
		})
	})
})
