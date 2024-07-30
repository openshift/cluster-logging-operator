package api

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/filters"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#ConvertLoggingToObservability", func() {

	Context("ClusterLogging instance only", func() {
		var (
			loggingCLSpec *logging.ClusterLoggingSpec
			outputName    string
		)
		BeforeEach(func() {
			loggingCLSpec = &logging.ClusterLoggingSpec{}
		})

		It("should convert to observability.ClusterLogForwarder with default elasticsearch output and pipeline", func() {
			outputName = "default-elasticsearch"
			esURL := "https://elasticsearch:9200"
			loggingCLSpec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}

			expObsClfSpec := &obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: constants.CollectorServiceAccountName,
				},
				Outputs: []obs.OutputSpec{
					{
						Name: outputName,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: esURL,
							},
							Version: 6,
							Index:   `{.log_type||"none"}`,
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
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name:       outputName + "-pipeline",
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{outputName},
					},
				},
			}

			actObsClfSpec := convertLegacyClusterLogging(loggingCLSpec)
			Expect(actObsClfSpec).To(Equal(expObsClfSpec))
		})
		It("should convert to observability.ClusterLogForwarder with default lokistack output and pipeline", func() {
			outputName = "default-lokistack"
			loggingCLSpec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeLokiStack,
				LokiStack: logging.LokiStackStoreSpec{
					Name: "my-lokistack",
				},
			}

			expObsClfSpec := &obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: constants.CollectorServiceAccountName,
				},
				Outputs: []obs.OutputSpec{
					{
						Name: outputName,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Namespace: constants.OpenshiftNS,
								Name:      "my-lokistack",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromSecret,
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
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name:       outputName + "-pipeline",
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{outputName},
					},
				},
			}

			actObsClfSpec := convertLegacyClusterLogging(loggingCLSpec)
			Expect(actObsClfSpec).To(Equal(expObsClfSpec))
		})

	})

	Context("convertClusterLogForwarder", func() {
		const (
			url = "https://0.0.0.0:9000"
		)
		var (
			k8sClient client.Client
			esSecret  = &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "es-secret",
					Namespace: "openshift-logging",
				},
				Data: map[string][]byte{
					constants.ClientUsername:     []byte("testuser"),
					constants.ClientPassword:     []byte("testpass"),
					constants.ClientPrivateKey:   []byte("akey"),
					constants.ClientCertKey:      []byte("acert"),
					constants.TrustedCABundleKey: []byte("aca"),
				},
			}

			cwSecret = &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cw-secret",
					Namespace: "openshift-logging",
				},
				Data: map[string][]byte{
					constants.AWSAccessKeyID:     []byte("accesskey"),
					constants.AWSSecretAccessKey: []byte("secretkey"),
				},
			}

			outputSecrets = map[string]*corev1.Secret{"es-out": esSecret, "cw": cwSecret}
		)

		DescribeTable("ConvertLoggingToObservability", func(expObsVisit func(obsClf *obs.ClusterLogForwarder), clVisit func(cl *logging.ClusterLogging) *logging.ClusterLogging, clfVisit func(clf *logging.ClusterLogForwarder)) {
			k8sClient = fake.NewClientBuilder().WithObjects(esSecret, cwSecret).Build()
			loggingCl := &logging.ClusterLogging{
				ObjectMeta: v1.ObjectMeta{
					Name:      constants.SingletonName,
					Namespace: constants.OpenshiftNS,
				},
				Spec: logging.ClusterLoggingSpec{},
			}
			loggingCl = clVisit(loggingCl)

			loggingClf := &logging.ClusterLogForwarder{
				ObjectMeta: v1.ObjectMeta{
					Name:      constants.SingletonName,
					Namespace: constants.OpenshiftNS,
				},
				Spec: logging.ClusterLogForwarderSpec{
					ServiceAccountName: constants.CollectorServiceAccountName,
					Inputs: []logging.InputSpec{
						{
							Name: "foo-app",
							Application: &logging.Application{
								Includes: []logging.NamespaceContainerSpec{
									{
										Container: "foo",
										Namespace: "bar",
									},
								},
							},
						},
					},
					Outputs: []logging.OutputSpec{
						{
							Name: "es-out",
							Type: logging.OutputTypeElasticsearch,
							URL:  url,
							Secret: &logging.OutputSecretSpec{
								Name: "es-secret",
							},
							TLS: &logging.OutputTLSSpec{
								InsecureSkipVerify: true,
							},
						},
						{
							Name: "cw",
							Type: logging.OutputTypeCloudwatch,
							OutputTypeSpec: logging.OutputTypeSpec{
								Cloudwatch: &logging.Cloudwatch{
									GroupBy: logging.LogGroupByLogType,
									Region:  "us-west-1",
								},
							},
							Secret: &logging.OutputSecretSpec{
								Name: "cw-secret",
							},
						},
						{
							Name: "my-http",
							Type: logging.OutputTypeHttp,
							URL:  url,
							OutputTypeSpec: logging.OutputTypeSpec{
								Http: &logging.Http{
									Method:  "POST",
									Headers: map[string]string{"foo": "bar"},
								},
							},
							Tuning: &logging.OutputTuningSpec{
								Delivery: logging.OutputDeliveryModeAtLeastOnce,
							},
						},
					},
					Filters: []logging.FilterSpec{
						{
							Name: "my-prune",
							Type: logging.FilterPrune,
							FilterTypeSpec: logging.FilterTypeSpec{
								PruneFilterSpec: &logging.PruneFilterSpec{
									In: []string{"foo", "bar"},
								},
							},
						},
					},
					Pipelines: []logging.PipelineSpec{
						{
							Name:       "app-logs",
							InputRefs:  []string{logging.InputNameApplication},
							OutputRefs: []string{"es-out"},
							Labels:     map[string]string{"foo": "bar"},
						},
						{
							Name:                  "custom-app",
							InputRefs:             []string{"foo-app", logging.InputNameAudit},
							OutputRefs:            []string{"cw", "my-http"},
							FilterRefs:            []string{"my-prune"},
							DetectMultilineErrors: true,
						},
					},
				},
			}

			clfVisit(loggingClf)

			expObsClf := &obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					ServiceAccount: obs.ServiceAccount{
						Name: constants.CollectorServiceAccountName,
					},
					Inputs: []obs.InputSpec{
						{
							Name: "foo-app",
							Type: obs.InputTypeApplication,
							Application: &obs.Application{
								Includes: []obs.NamespaceContainerSpec{
									{
										Container: "foo",
										Namespace: "bar",
									},
								},
							},
						},
					},
					Outputs: []obs.OutputSpec{
						{
							Name: "es-out",
							Type: obs.OutputTypeElasticsearch,
							Elasticsearch: &obs.Elasticsearch{
								URLSpec: obs.URLSpec{
									URL: url,
								},
								Version: 8,
								Authentication: &obs.HTTPAuthentication{
									Username: &obs.SecretReference{
										Key:        constants.ClientUsername,
										SecretName: "es-secret",
									},
									Password: &obs.SecretReference{
										Key:        constants.ClientPassword,
										SecretName: "es-secret",
									},
								},
								Index: `{.log_type||"none"}`,
							},
							TLS: &obs.OutputTLSSpec{
								InsecureSkipVerify: true,
								TLSSpec: obs.TLSSpec{
									CA: &obs.ValueReference{
										Key:        constants.TrustedCABundleKey,
										SecretName: "es-secret",
									},
									Certificate: &obs.ValueReference{
										Key:        constants.ClientCertKey,
										SecretName: "es-secret",
									},
									Key: &obs.SecretReference{
										Key:        constants.ClientPrivateKey,
										SecretName: "es-secret",
									},
								},
							},
						},
						{
							Name: "cw",
							Type: logging.OutputTypeCloudwatch,
							Cloudwatch: &obs.Cloudwatch{
								GroupName: `{.openshift.cluster_id||"none"}.{.log_type||"none"}`,
								Region:    "us-west-1",
								Authentication: &obs.CloudwatchAuthentication{
									Type: obs.CloudwatchAuthTypeAccessKey,
									AWSAccessKey: &obs.CloudwatchAWSAccessKey{
										KeyID: &obs.SecretReference{
											Key:        constants.AWSAccessKeyID,
											SecretName: "cw-secret",
										},
										KeySecret: &obs.SecretReference{
											Key:        constants.AWSSecretAccessKey,
											SecretName: "cw-secret",
										},
									},
								},
							},
						},
						{
							Name: "my-http",
							Type: logging.OutputTypeHttp,
							HTTP: &obs.HTTP{
								URLSpec: obs.URLSpec{
									URL: url,
								},
								Method:  "POST",
								Headers: map[string]string{"foo": "bar"},
								Tuning: &obs.HTTPTuningSpec{
									BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
										Delivery: obs.DeliveryModeAtLeastOnce,
									},
								},
							},
						},
					},
					Filters: []obs.FilterSpec{
						{
							Name: "my-prune",
							Type: obs.FilterTypePrune,
							PruneFilterSpec: &obs.PruneFilterSpec{
								In: []obs.FieldPath{"foo", "bar"},
							},
						},
						{
							Name:            "filter-app-logs-" + filters.OpenshiftLabelsFilterName,
							Type:            obs.FilterTypeOpenshiftLabels,
							OpenShiftLabels: map[string]string{"foo": "bar"},
						},
						{
							Name: filters.DetectMultilineErrorFilterName,
							Type: obs.FilterTypeDetectMultiline,
						},
					},
					Pipelines: []obs.PipelineSpec{
						{
							Name:       "app-logs",
							InputRefs:  []string{logging.InputNameApplication},
							OutputRefs: []string{"es-out"},
						},
						{
							Name:       "custom-app",
							InputRefs:  []string{"foo-app", logging.InputNameAudit},
							OutputRefs: []string{"cw", "my-http"},
							FilterRefs: []string{"my-prune"},
						},
					},
				},
			}

			expObsVisit(expObsClf)

			actObsClfSpec := ConvertLoggingToObservability(k8sClient, loggingCl, loggingClf, outputSecrets)
			Expect(actObsClfSpec.Spec.ServiceAccount).To(Equal(expObsClf.Spec.ServiceAccount))
			Expect(actObsClfSpec.Spec.Collector).To(Equal(expObsClf.Spec.Collector))
			Expect(actObsClfSpec.Spec.Inputs).To(Equal(expObsClf.Spec.Inputs))
			Expect(actObsClfSpec.Spec.Outputs).To(Equal(expObsClf.Spec.Outputs))
			Expect(actObsClfSpec.Spec.Filters).To(Equal(expObsClf.Spec.Filters))
		},
			Entry("with legacy ClusterLogging & ClusterLogForwarder",
				func(expObsClf *obs.ClusterLogForwarder) {
					expObsClf.Spec.Collector = &obs.CollectorSpec{
						NodeSelector: map[string]string{"foo": "bar"},
					}
					expObsClf.Spec.Outputs = append(expObsClf.Spec.Outputs, obs.OutputSpec{
						Name: "default-elasticsearch",
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: "https://elasticsearch:9200",
							},
							Version: 6,
							Index:   `{.log_type||"none"}`,
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
					})
					expObsClf.Spec.Pipelines[0].OutputRefs = append(expObsClf.Spec.Pipelines[0].OutputRefs, "default-elasticsearch")
				},
				func(cl *logging.ClusterLogging) *logging.ClusterLogging {
					cl.Spec = logging.ClusterLoggingSpec{
						LogStore: &logging.LogStoreSpec{
							Type: logging.LogStoreTypeElasticsearch,
						},
						Collection: &logging.CollectionSpec{
							CollectorSpec: logging.CollectorSpec{
								NodeSelector: map[string]string{"foo": "bar"},
							},
						},
					}
					return cl
				},
				func(clf *logging.ClusterLogForwarder) {
					clf.Spec.Pipelines[0].OutputRefs = append(clf.Spec.Pipelines[0].OutputRefs, "default")
				}),
			Entry("with custom logging ClusterLogForwarder only",
				func(expObsClf *obs.ClusterLogForwarder) {
					expObsClf.Name = "custom-clf"
					expObsClf.Spec.ServiceAccount.Name = "test-sa"
				},
				func(cl *logging.ClusterLogging) *logging.ClusterLogging {
					return nil
				},
				func(clf *logging.ClusterLogForwarder) {
					clf.Name = "custom-clf"
					clf.Spec.ServiceAccountName = "test-sa"
				}),
			Entry("with custom logging ClusterLogging & logging ClusterLogForwarder",
				func(expObsClf *obs.ClusterLogForwarder) {
					expObsClf.Name = "cl-clf"
					expObsClf.Spec.ServiceAccount.Name = "my-sa"
					expObsClf.Spec.Collector = &obs.CollectorSpec{
						NodeSelector: map[string]string{"foo": "bar"},
						Tolerations: []corev1.Toleration{
							{
								Key:      "tol1",
								Operator: corev1.TolerationOpEqual,
								Value:    "val1",
								Effect:   corev1.TaintEffectNoSchedule,
							},
						},
					}
				},
				func(cl *logging.ClusterLogging) *logging.ClusterLogging {
					cl.Name = "cl-clf"
					cl.Spec.Collection = &logging.CollectionSpec{
						CollectorSpec: logging.CollectorSpec{
							NodeSelector: map[string]string{"foo": "bar"},
							Tolerations: []corev1.Toleration{
								{
									Key:      "tol1",
									Operator: corev1.TolerationOpEqual,
									Value:    "val1",
									Effect:   corev1.TaintEffectNoSchedule,
								},
							},
						},
					}
					return cl
				},
				func(clf *logging.ClusterLogForwarder) {
					clf.Spec.ServiceAccountName = "my-sa"
					clf.Name = "cl-clf"
				}))
	})
})
