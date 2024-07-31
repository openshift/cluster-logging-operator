package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/config/v1"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("#MapOutputs", func() {
	It("should map logging outputs to observability outputs", func() {
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
					constants.SplunkHECTokenKey:  []byte("hec-token"),
					constants.TrustedCABundleKey: []byte("ca"),
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
					Index:   `{.log_type||"none"}`,
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
					Index: "app",
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
				TLS: &obs.OutputTLSSpec{
					TLSSpec: obs.TLSSpec{
						CA: &obs.ValueReference{
							Key:        constants.TrustedCABundleKey,
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
					LabelKeys: []string{"foo", "bar"},
					Authentication: &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromSecret,
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

		obsClfSpecOutputs := ConvertOutputs(loggingClfSpec, secrets)
		Expect(obsClfSpecOutputs).To(Equal(expObsClfOutputs))
	})

})
