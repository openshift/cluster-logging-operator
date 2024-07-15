package api

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/config/v1"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#ConvertOutputs", func() {
	Context("default output", func() {
		It("should generate default elasticsearch output based on logstoreSpec", func() {
			logStoreSpec := logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}
			expEsOut := &obs.OutputSpec{
				Name: "default-elasticsearch",
				Type: obs.OutputTypeElasticsearch,
				Elasticsearch: &obs.Elasticsearch{
					URLSpec: obs.URLSpec{
						URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
					},
					Version: 6,
					IndexSpec: obs.IndexSpec{
						Index: "{{.log_type}}",
					},
				},
				TLS: &obs.OutputTLSSpec{
					TLSSpec: obs.TLSSpec{
						CA: &obs.ValueReference{
							Key:        constants.TrustedCABundleKey,
							SecretName: constants.ElasticsearchName,
						},
						Certificate: &obs.ValueReference{
							Key:        constants.ClientCertKey,
							SecretName: constants.ElasticsearchName,
						},
						Key: &obs.SecretReference{
							Key:        constants.ClientPrivateKey,
							SecretName: constants.ElasticsearchName,
						},
					},
				},
			}
			Expect(generateDefaultOutput(&logStoreSpec)).To(Equal(expEsOut))
		})

		It("should generate default lokistack output", func() {
			logStoreSpec := logging.LogStoreSpec{
				Type: logging.LogStoreTypeLokiStack,
				LokiStack: logging.LokiStackStoreSpec{
					Name: "my-lokistack",
				},
			}
			expLokiStackOut := &obs.OutputSpec{
				Name: "default-lokistack",
				Type: obs.OutputTypeLokiStack,
				LokiStack: &obs.LokiStack{
					Target: obs.LokiStackTarget{
						Name:      "my-lokistack",
						Namespace: constants.OpenshiftNS,
					},
					Authentication: &obs.LokiStackAuthentication{
						Token: &obs.BearerToken{
							Secret: &obs.BearerTokenSecretKey{
								Name: constants.LogCollectorToken,
								Key:  constants.BearerTokenFileKey,
							},
						},
					},
				},
				TLS: &obs.OutputTLSSpec{
					TLSSpec: obs.TLSSpec{
						CA: &obs.ValueReference{
							Key:        "service-ca.crt",
							SecretName: constants.LogCollectorToken,
						},
					},
				},
			}
			Expect(generateDefaultOutput(&logStoreSpec)).To(Equal(expLokiStackOut))
		})
	})
	Context("output helper functions", func() {
		const secretName = "my-secret"
		var (
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
					constants.ClientUsername:     []byte("username"),
					constants.ClientPassword:     []byte("password"),
					constants.BearerTokenFileKey: []byte("token"),
				},
			}
		)

		It("should map logging output TLS to observability TLS", func() {
			loggingTLS := &logging.OutputTLSSpec{
				InsecureSkipVerify: true,
				TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
					Type:   openshiftv1.TLSProfileType("foo"),
					Modern: &openshiftv1.ModernTLSProfile{},
				},
			}

			expTLS := &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
				TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
					Type:   openshiftv1.TLSProfileType("foo"),
					Modern: &openshiftv1.ModernTLSProfile{},
				},
				TLSSpec: obs.TLSSpec{
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: secretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: secretName,
					},
					KeyPassphrase: &obs.SecretReference{
						Key:        constants.Passphrase,
						SecretName: secretName,
					},
				},
			}
			actOutTLS := mapOutputTls(loggingTLS, secret)
			Expect(actOutTLS).To(Equal(expTLS))
		})

		It("should map logging base output tuning to observability base tuning", func() {
			loggingOutputBaseTune := logging.OutputTuningSpec{
				Delivery:         logging.OutputDeliveryModeAtLeastOnce,
				MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
				MinRetryDuration: utils.GetPtr(time.Duration(1)),
				MaxRetryDuration: utils.GetPtr(time.Duration(5)),
			}
			obsOutputBaseTune := &obs.BaseOutputTuningSpec{
				Delivery:         obs.DeliveryModeAtLeastOnce,
				MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
				MinRetryDuration: utils.GetPtr(time.Duration(1)),
				MaxRetryDuration: utils.GetPtr(time.Duration(5)),
			}
			actualObsBaseTune := mapBaseOutputTuning(loggingOutputBaseTune)

			Expect(actualObsBaseTune).To(Equal(obsOutputBaseTune))
		})

		It("should map logging HTTPAuthentication to observability HTTPAuthentication", func() {
			obsHttpAuth := &obs.HTTPAuthentication{
				Username: &obs.SecretReference{
					Key:        constants.ClientUsername,
					SecretName: secretName,
				},
				Password: &obs.SecretReference{
					Key:        constants.ClientPassword,
					SecretName: secretName,
				},
				Token: &obs.BearerToken{
					Secret: &obs.BearerTokenSecretKey{
						Name: secretName,
						Key:  constants.BearerTokenFileKey,
					},
				},
			}
			actualHTTPAuth := mapHTTPAuth(secret)
			Expect(actualHTTPAuth).To(Equal(obsHttpAuth))
		})
	})

	Context("output type specs", func() {
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
		It("should map logging.AzureMonitor to obs.AzureMonitor", func() {
			secret.Data[constants.SharedKey] = []byte("shared-key")
			loggingOutSpec := logging.OutputSpec{
				OutputTypeSpec: logging.OutputTypeSpec{
					AzureMonitor: &logging.AzureMonitor{
						CustomerId:      "cust",
						LogType:         "app",
						AzureResourceId: "my-id",
						Host:            "my-host",
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery: logging.OutputDeliveryModeAtLeastOnce,
				},
			}
			expAzMon := &obs.AzureMonitor{
				CustomerId:      "cust",
				LogType:         "app",
				AzureResourceId: "my-id",
				Host:            "my-host",
				Authentication: &obs.AzureMonitorAuthentication{
					SharedKey: &obs.SecretReference{
						Key:        constants.SharedKey,
						SecretName: secretName,
					},
				},
				Tuning: &obs.BaseOutputTuningSpec{
					Delivery: obs.DeliveryModeAtLeastOnce,
				},
			}
			Expect(mapAzureMonitor(loggingOutSpec, secret)).To(Equal(expAzMon))
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
				GroupName: "prefix.{{.log_type}}",
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
			Expect(mapCloudwatch(loggingOutSpec, secret, "")).To(Equal(expectedCWOut))
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
				GroupName: "prefix.{{.log_type}}",
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
							Secret: &obs.BearerTokenSecretKey{
								Name: secretName,
								Key:  constants.BearerTokenFileKey,
							},
						},
					},
				},
			}
			Expect(mapCloudwatch(loggingOutSpec, secret, "")).To(Equal(expectedCWOut))
		})
		It("should map logging.Elasticsearch to obs.Elasticsearch", func() {
			secret.Data[constants.ClientUsername] = []byte("user")
			secret.Data[constants.ClientPassword] = []byte("pass")

			loggingOutSpec := logging.OutputSpec{
				URL: url,
				OutputTypeSpec: logging.OutputTypeSpec{
					Elasticsearch: &logging.Elasticsearch{
						Version: 8,
						ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
							StructuredTypeKey:  ".namespace",
							StructuredTypeName: "structName",
						},
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery:    logging.OutputDeliveryModeAtLeastOnce,
					Compression: "gzip",
				},
			}

			expObsElastic := &obs.Elasticsearch{
				URLSpec: obs.URLSpec{
					URL: url,
				},
				Version: 8,
				IndexSpec: obs.IndexSpec{
					Index: ".namespace",
				},
				Authentication: &obs.HTTPAuthentication{
					Username: &obs.SecretReference{
						Key:        constants.ClientUsername,
						SecretName: secretName,
					},
					Password: &obs.SecretReference{
						Key:        constants.ClientPassword,
						SecretName: secretName,
					},
				},
				Tuning: &obs.ElasticsearchTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						Delivery: obs.DeliveryModeAtLeastOnce,
					},
					Compression: "gzip",
				},
			}

			Expect(mapElasticsearch(loggingOutSpec, secret)).To(Equal(expObsElastic))
		})
		It("should map logging.GoogleCloudLogging to obs.GoogleCloudLogging", func() {
			secret.Data[gcl.GoogleApplicationCredentialsKey] = []byte("google.json")

			loggingOutSpec := logging.OutputSpec{
				OutputTypeSpec: logging.OutputTypeSpec{
					GoogleCloudLogging: &logging.GoogleCloudLogging{
						BillingAccountID: "foo",

						LogID: "baz",
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery:         logging.OutputDeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				},
			}

			expObsGCP := &obs.GoogleCloudLogging{
				ID: obs.GoogleGloudLoggingID{
					Type:  obs.GoogleCloudLoggingIDTypeBillingAccount,
					Value: "foo",
				},
				LogID: "baz",
				Authentication: &obs.GoogleCloudLoggingAuthentication{
					Credentials: &obs.SecretReference{
						Key:        gcl.GoogleApplicationCredentialsKey,
						SecretName: secretName,
					},
				},
				Tuning: &obs.GoogleCloudLoggingTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						Delivery:         obs.DeliveryModeAtLeastOnce,
						MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
						MinRetryDuration: utils.GetPtr(time.Duration(1)),
						MaxRetryDuration: utils.GetPtr(time.Duration(5)),
					},
				},
			}

			Expect(mapGoogleCloudLogging(loggingOutSpec, secret)).To(Equal(expObsGCP))

		})
		It("should map logging.HTTP to obs.HTTP", func() {
			secret.Data[constants.ClientUsername] = []byte("user")
			secret.Data[constants.ClientPassword] = []byte("pass")

			headers := map[string]string{"k1": "v1", "k2": "v2"}

			loggingOutSpec := logging.OutputSpec{
				URL: url,
				OutputTypeSpec: logging.OutputTypeSpec{
					Http: &logging.Http{
						Headers: headers,
						Method:  "POST",
						Timeout: 100,
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery:         logging.OutputDeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
					Compression:      "gzip",
				},
			}

			expObsHTTP := &obs.HTTP{
				URLSpec: obs.URLSpec{
					URL: url,
				},
				Headers: headers,
				Method:  "POST",
				Timeout: 100,
				Authentication: &obs.HTTPAuthentication{
					Username: &obs.SecretReference{
						Key:        constants.ClientUsername,
						SecretName: secretName,
					},
					Password: &obs.SecretReference{
						Key:        constants.ClientPassword,
						SecretName: secretName,
					},
				},
				Tuning: &obs.HTTPTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						Delivery:         obs.DeliveryModeAtLeastOnce,
						MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
						MinRetryDuration: utils.GetPtr(time.Duration(1)),
						MaxRetryDuration: utils.GetPtr(time.Duration(5)),
					},
					Compression: "gzip",
				},
			}

			Expect(mapHTTP(loggingOutSpec, secret)).To(Equal(expObsHTTP))
		})
		It("should map logging.Kafka to obs.Kafka", func() {
			secret.Data[constants.ClientUsername] = []byte("user")
			secret.Data[constants.ClientPassword] = []byte("pass")
			secret.Data[constants.SASLMechanisms] = []byte("SCRAM-SHA-256")

			loggingOutSpec := logging.OutputSpec{
				URL: url,
				OutputTypeSpec: logging.OutputTypeSpec{
					Kafka: &logging.Kafka{
						Topic:   "foo",
						Brokers: []string{"foo", "bar"},
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery:    logging.OutputDeliveryModeAtLeastOnce,
					MaxWrite:    utils.GetPtr(resource.MustParse("100m")),
					Compression: "zstd",
				},
			}

			expObsKafka := &obs.Kafka{
				URLSpec: obs.URLSpec{
					URL: url,
				},
				Topic:   "foo",
				Brokers: []string{"foo", "bar"},
				Authentication: &obs.KafkaAuthentication{
					SASL: &obs.SASLAuthentication{
						Username: &obs.SecretReference{
							Key:        constants.ClientUsername,
							SecretName: secretName,
						},
						Password: &obs.SecretReference{
							Key:        constants.ClientPassword,
							SecretName: secretName,
						},
						Mechanism: "SCRAM-SHA-256",
					},
				},
				Tuning: &obs.KafkaTuningSpec{
					Delivery:    obs.DeliveryModeAtLeastOnce,
					MaxWrite:    utils.GetPtr(resource.MustParse("100m")),
					Compression: "zstd",
				},
			}

			Expect(mapKafka(loggingOutSpec, secret)).To(Equal(expObsKafka))
		})
		It("should map logging.Loki to obs.Loki", func() {
			secret.Data[constants.BearerTokenFileKey] = []byte("token")
			loggingOutSpec := logging.OutputSpec{
				URL: url,
				OutputTypeSpec: logging.OutputTypeSpec{
					Loki: &logging.Loki{
						TenantKey: "app",
						LabelKeys: []string{"foo", "bar"},
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery:         logging.OutputDeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
					Compression:      "snappy",
				},
			}
			expObsLoki := &obs.Loki{
				URLSpec: obs.URLSpec{
					URL: url,
				},
				TenantKey: "app",
				LabelKeys: []string{"foo", "bar"},
				Tuning: &obs.LokiTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						Delivery:         obs.DeliveryModeAtLeastOnce,
						MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
						MinRetryDuration: utils.GetPtr(time.Duration(1)),
						MaxRetryDuration: utils.GetPtr(time.Duration(5)),
					},
					Compression: "snappy",
				},
				Authentication: &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						Secret: &obs.BearerTokenSecretKey{
							Name: secretName,
							Key:  constants.BearerTokenFileKey,
						},
					},
				},
			}

			Expect(mapLoki(loggingOutSpec, secret)).To(Equal(expObsLoki))
		})
		It("should map logging.Splunk to obs.Splunk", func() {
			secret.Data[constants.SplunkHECTokenKey] = []byte("hec-token")
			loggingOutSpec := logging.OutputSpec{
				URL: url,
				OutputTypeSpec: logging.OutputTypeSpec{
					Splunk: &logging.Splunk{
						IndexKey: ".bar",
					},
				},
				Tuning: &logging.OutputTuningSpec{
					Delivery:         logging.OutputDeliveryModeAtLeastOnce,
					MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
					MinRetryDuration: utils.GetPtr(time.Duration(1)),
					MaxRetryDuration: utils.GetPtr(time.Duration(5)),
				},
			}

			expObsSplunk := &obs.Splunk{
				URLSpec: obs.URLSpec{
					URL: url,
				},
				IndexSpec: obs.IndexSpec{
					Index: ".bar",
				},
				Tuning: &obs.SplunkTuningSpec{
					BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
						Delivery:         obs.DeliveryModeAtLeastOnce,
						MaxWrite:         utils.GetPtr(resource.MustParse("100m")),
						MinRetryDuration: utils.GetPtr(time.Duration(1)),
						MaxRetryDuration: utils.GetPtr(time.Duration(5)),
					},
				},
				Authentication: &obs.SplunkAuthentication{
					Token: &obs.SecretReference{
						Key:        constants.SplunkHECTokenKey,
						SecretName: secretName,
					},
				},
			}

			Expect(mapSplunk(loggingOutSpec, secret)).To(Equal(expObsSplunk))
		})
		It("should map logging.Syslog to obs.Syslog", func() {
			loggingOutSpec := logging.OutputSpec{
				URL: url,
				OutputTypeSpec: logging.OutputTypeSpec{
					Syslog: &logging.Syslog{
						RFC:        "RFC3164",
						Severity:   "error",
						Facility:   "foo",
						PayloadKey: "bar",
						AppName:    "app",
						ProcID:     "123",
						MsgID:      "12345",
					},
				},
			}

			expObsSyslog := &obs.Syslog{
				URLSpec: obs.URLSpec{
					URL: url,
				},
				RFC:        "RFC3164",
				Severity:   "error",
				Facility:   "foo",
				PayloadKey: "bar",
				AppName:    "app",
				ProcID:     "123",
				MsgID:      "12345",
			}

			Expect(mapSyslog(loggingOutSpec)).To(Equal(expObsSyslog))
		})
	})

	It("should convert logging outputs to observability outputs", func() {
		var (
			esSecret = runtime.NewSecret(
				"test-namespace",
				"es-secret",
				map[string][]byte{
					constants.ClientUsername: []byte("user"),
					constants.ClientPassword: []byte("pass"),
				})

			splunkSecret = runtime.NewSecret(
				"test-namespace",
				"splunk-secret",
				map[string][]byte{
					constants.SplunkHECTokenKey: []byte("hec-token"),
				},
			)

			lokiSecret = runtime.NewSecret(
				"test-namespace",
				"loki-secret",
				map[string][]byte{
					constants.BearerTokenFileKey: []byte("token"),
					constants.TrustedCABundleKey: []byte("ca"),
					constants.ClientCertKey:      []byte("cert"),
					constants.ClientPrivateKey:   []byte("privatekey"),
				},
			)

			azureSecret = runtime.NewSecret(
				"test-namespace",
				"azure-secret",
				map[string][]byte{
					constants.SharedKey:          []byte("shared"),
					constants.TrustedCABundleKey: []byte("ca"),
					constants.ClientCertKey:      []byte("cert"),
					constants.ClientPrivateKey:   []byte("privatekey"),
				},
			)

			secrets = map[string]*corev1.Secret{
				"es-out":     esSecret,
				"splunk-out": splunkSecret,
				"loki-out":   lokiSecret,
				"azure-out":  azureSecret,
			}

			url = "https://0.0.0.0"
		)

		loggingClfSpec := &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Name: "es-out",
					Type: logging.OutputTypeElasticsearch,
					URL:  url,
					Tuning: &logging.OutputTuningSpec{
						Compression: "gzip",
					},
					Secret: &logging.OutputSecretSpec{
						Name: esSecret.Name,
					},
				},
				{
					Name: "splunk-out",
					Type: logging.OutputTypeSplunk,
					URL:  url,
					Tuning: &logging.OutputTuningSpec{
						Delivery: logging.OutputDeliveryModeAtMostOnce,
					},
					Secret: &logging.OutputSecretSpec{
						Name: splunkSecret.Name,
					},
					OutputTypeSpec: logging.OutputTypeSpec{
						Splunk: &logging.Splunk{
							IndexName: "app",
						},
					},
				},
				{
					Name: "loki-out",
					Type: logging.OutputTypeLoki,
					URL:  url,
					Tuning: &logging.OutputTuningSpec{
						Compression: "gzip",
					},
					Secret: &logging.OutputSecretSpec{
						Name: lokiSecret.Name,
					},
					OutputTypeSpec: logging.OutputTypeSpec{
						Loki: &logging.Loki{
							LabelKeys: []string{"foo", "bar"},
						},
					},
					TLS: &logging.OutputTLSSpec{
						InsecureSkipVerify: true,
					},
				},
				{
					Name: "azure-out",
					Type: logging.OutputTypeAzureMonitor,
					URL:  url,
					Tuning: &logging.OutputTuningSpec{
						Delivery: logging.OutputDeliveryModeAtLeastOnce,
					},
					Secret: &logging.OutputSecretSpec{
						Name: azureSecret.Name,
					},
					TLS: &logging.OutputTLSSpec{
						TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
							Old: &openshiftv1.OldTLSProfile{},
						},
					},
					OutputTypeSpec: logging.OutputTypeSpec{
						AzureMonitor: &logging.AzureMonitor{
							Host:            "foo",
							LogType:         "app",
							CustomerId:      "bar",
							AzureResourceId: "baz",
						},
					},
				},
			},
		}

		expObsClfOutputs := []obs.OutputSpec{
			{
				Name: "es-out",
				Type: obs.OutputTypeElasticsearch,
				Elasticsearch: &obs.Elasticsearch{
					URLSpec: obs.URLSpec{
						URL: url,
					},
					Version: 8,
					IndexSpec: obs.IndexSpec{
						Index: "{{.log_type}}",
					},
					Authentication: &obs.HTTPAuthentication{
						Username: &obs.SecretReference{
							Key:        constants.ClientUsername,
							SecretName: esSecret.Name,
						},
						Password: &obs.SecretReference{
							Key:        constants.ClientPassword,
							SecretName: esSecret.Name,
						},
					},
					Tuning: &obs.ElasticsearchTuningSpec{
						Compression: "gzip",
					},
				},
			},
			{
				Name: "splunk-out",
				Type: obs.OutputTypeSplunk,
				Splunk: &obs.Splunk{
					URLSpec: obs.URLSpec{
						URL: url,
					},
					IndexSpec: obs.IndexSpec{
						Index: "app",
					},
					Tuning: &obs.SplunkTuningSpec{
						BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
							Delivery: obs.DeliveryModeAtMostOnce,
						},
					},
					Authentication: &obs.SplunkAuthentication{
						Token: &obs.SecretReference{
							Key:        constants.SplunkHECTokenKey,
							SecretName: splunkSecret.Name,
						},
					},
				},
			},
			{
				Name: "loki-out",
				Type: obs.OutputTypeLoki,
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: url,
					},
					TenantKey: "{{.log_type}}",
					LabelKeys: []string{"foo", "bar"},
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							Secret: &obs.BearerTokenSecretKey{
								Name: lokiSecret.Name,
								Key:  constants.BearerTokenFileKey,
							},
						},
					},
					Tuning: &obs.LokiTuningSpec{
						Compression: "gzip",
					},
				},
				TLS: &obs.OutputTLSSpec{
					InsecureSkipVerify: true,
					TLSSpec: obs.TLSSpec{
						Certificate: &obs.ValueReference{
							Key:        constants.ClientCertKey,
							SecretName: lokiSecret.Name,
						},
						Key: &obs.SecretReference{
							Key:        constants.ClientPrivateKey,
							SecretName: lokiSecret.Name,
						},
						CA: &obs.ValueReference{
							Key:        constants.TrustedCABundleKey,
							SecretName: lokiSecret.Name,
						},
					},
				},
			},
			{
				Name: "azure-out",
				Type: obs.OutputTypeAzureMonitor,
				AzureMonitor: &obs.AzureMonitor{
					Host:            "foo",
					LogType:         "app",
					CustomerId:      "bar",
					AzureResourceId: "baz",
					Authentication: &obs.AzureMonitorAuthentication{
						SharedKey: &obs.SecretReference{
							Key:        constants.SharedKey,
							SecretName: azureSecret.Name,
						},
					},
					Tuning: &obs.BaseOutputTuningSpec{
						Delivery: obs.DeliveryModeAtLeastOnce,
					},
				},
				TLS: &obs.OutputTLSSpec{
					TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
						Old: &openshiftv1.OldTLSProfile{},
					},
					TLSSpec: obs.TLSSpec{
						Certificate: &obs.ValueReference{
							Key:        constants.ClientCertKey,
							SecretName: azureSecret.Name,
						},
						Key: &obs.SecretReference{
							Key:        constants.ClientPrivateKey,
							SecretName: azureSecret.Name,
						},
						CA: &obs.ValueReference{
							Key:        constants.TrustedCABundleKey,
							SecretName: azureSecret.Name,
						},
					},
				},
			},
		}

		obsClfSpecOutputs := convertOutputs(loggingClfSpec, secrets)
		Expect(obsClfSpecOutputs).To(Equal(expObsClfOutputs))
	})

})
