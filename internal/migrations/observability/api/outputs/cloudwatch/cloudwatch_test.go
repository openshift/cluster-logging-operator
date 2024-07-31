package cloudwatch

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapCloudwatch", func() {
	const secretName = "my-secret"
	var (
		secret *corev1.Secret
		url    = "0.0.0.0:9200"
	)
	BeforeEach(func() {
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: "foo-space",
			},
			Data: map[string][]byte{
				constants.ClientCertKey:      []byte("cert"),
				constants.ClientPrivateKey:   []byte("privatekey"),
				constants.TrustedCABundleKey: []byte("cabundle"),
				constants.Passphrase:         []byte("pass"),
			},
		}
	})
	It("should map logging.Cloudwatch to obs.Cloudwatch with KeyId & Key Secret", func() {
		secret.Data[constants.AWSAccessKeyID] = []byte("accesskeyid")
		secret.Data[constants.AWSSecretAccessKey] = []byte("secretId")
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Cloudwatch: &logging.Cloudwatch{
					Region:      "us-west",
					GroupBy:     logging.LogGroupByLogType,
					GroupPrefix: utils.GetPtr("prefix"),
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:    logging.OutputDeliveryModeAtLeastOnce,
				Compression: "gzip",
			},
		}
		expectedCWOut := &obs.Cloudwatch{
			URL:       url,
			Region:    "us-west",
			GroupName: `prefix.{.log_type||"none"}`,
			Tuning: &obs.CloudwatchTuningSpec{
				Compression: "gzip",
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery: obs.DeliveryModeAtLeastOnce,
				},
			},
			Authentication: &obs.CloudwatchAuthentication{
				Type: obs.CloudwatchAuthTypeAccessKey,
				AWSAccessKey: &obs.CloudwatchAWSAccessKey{
					KeyID: &obs.SecretReference{
						Key:        constants.AWSAccessKeyID,
						SecretName: secretName,
					},
					KeySecret: &obs.SecretReference{
						Key:        constants.AWSSecretAccessKey,
						SecretName: secretName,
					},
				},
			},
		}
		Expect(MapCloudwatch(loggingOutSpec, secret)).To(Equal(expectedCWOut))
	})
	It("should map logging.Cloudwatch to obs.Cloudwatch with role_arn & token", func() {
		secret.Data[constants.AWSWebIdentityRoleKey] = []byte("test-role-arn")
		secret.Data[constants.BearerTokenFileKey] = []byte("my-token")
		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Cloudwatch: &logging.Cloudwatch{
					Region:      "us-west",
					GroupBy:     logging.LogGroupByLogType,
					GroupPrefix: utils.GetPtr("prefix"),
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:    logging.OutputDeliveryModeAtLeastOnce,
				Compression: "gzip",
			},
		}
		expectedCWOut := &obs.Cloudwatch{
			URL:       url,
			Region:    "us-west",
			GroupName: `prefix.{.log_type||"none"}`,
			Tuning: &obs.CloudwatchTuningSpec{
				Compression: "gzip",
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery: obs.DeliveryModeAtLeastOnce,
				},
			},
			Authentication: &obs.CloudwatchAuthentication{
				Type: obs.CloudwatchAuthTypeIAMRole,
				IAMRole: &obs.CloudwatchIAMRole{
					RoleARN: &obs.SecretReference{
						Key:        constants.AWSWebIdentityRoleKey,
						SecretName: secretName,
					},
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Name: secretName,
							Key:  constants.BearerTokenFileKey,
						},
					},
				},
			},
		}
		Expect(MapCloudwatch(loggingOutSpec, secret)).To(Equal(expectedCWOut))
	})
})
