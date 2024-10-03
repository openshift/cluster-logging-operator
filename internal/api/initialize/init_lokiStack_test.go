package initialize

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MigrateLokiStack", func() {
	const (
		esOut             = "es-out"
		lokistackOut      = "lokistack-out"
		lokistackOtlpOut  = "lokistack-otlp-out"
		lokistackTarget   = "test-lokistack"
		lokistackPipeline = "test-lokistack-pipeline"
		lokistackOutApp   = lokistackOut + "-" + string(obs.InputTypeApplication)
		lokistackOutAudit = lokistackOut + "-" + string(obs.InputTypeAudit)
		lokistackOutInfra = lokistackOut + "-" + string(obs.InputTypeInfrastructure)

		lokistackOtlpOutApp   = lokistackOtlpOut + "-" + string(obs.InputTypeApplication)
		lokistackOtlpOutAudit = lokistackOtlpOut + "-" + string(obs.InputTypeAudit)
		lokistackOtlpOutInfra = lokistackOtlpOut + "-" + string(obs.InputTypeInfrastructure)
	)

	var (
		spec    obs.ClusterLogForwarder
		initClf = func(outName string, isOtlp bool) obs.ClusterLogForwarder {
			obsClf := obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{
						{
							Name: outName,
							Type: obs.OutputTypeLokiStack,
							LokiStack: &obs.LokiStack{
								Target: obs.LokiStackTarget{
									Name:      lokistackTarget,
									Namespace: constants.OpenshiftNS,
								},
								Authentication: &obs.LokiStackAuthentication{
									Token: &obs.BearerToken{
										From: obs.BearerTokenFromServiceAccount,
									},
								},
							},
						},
					},
				},
			}

			if isOtlp {
				obsClf.Spec.Outputs[0].LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
			}

			return obsClf
		}
		esOutSpec = obs.OutputSpec{
			Name: esOut,
			Type: obs.OutputTypeElasticsearch,
			Elasticsearch: &obs.Elasticsearch{
				URLSpec: obs.URLSpec{
					URL: "https://my-elastic:9200",
				},
			},
		}
	)

	DescribeTable("migrate lokistack to loki outputs/pipelines", func(expSpec obs.ClusterLogForwarderSpec, visit func(spec *obs.ClusterLogForwarderSpec)) {
		clfSpec := initClf(lokistackOut, false)
		if visit != nil {
			visit(&clfSpec.Spec)
		}

		spec = MigrateLokiStack(clfSpec, utils.NoOptions)
		Expect(spec.Spec).To(Equal(expSpec))
	},
		Entry("single tenant, single lokistack output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOut},
					},
				}
			},
		),
		Entry("multiple tenants, single lokistack output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOutAudit},
					},
					{
						Name:       lokistackPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOutInfra},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOut},
					},
				}
			},
		),
		Entry("multiple tenants, single lokistack output, customized LabelKeys",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOutAudit},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							LabelKeys: []string{
								"log_type",
								"objectRef.apiGroup",
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = []obs.OutputSpec{
					{
						Name: lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      lokistackTarget,
								Namespace: constants.OpenshiftNS,
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							LabelKeys: &obs.LokiStackLabelKeys{
								Audit: &obs.LokiStackTenantLabelKeys{
									IgnoreGlobal: true,
									LabelKeys: []string{
										"log_type",
										"objectRef.apiGroup",
									},
								},
							},
						},
					},
				}
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name: lokistackPipeline,
						InputRefs: []string{
							string(obs.InputTypeApplication),
							string(obs.InputTypeAudit),
						},
						OutputRefs: []string{
							lokistackOut,
						},
					},
				}
			},
		),
		Entry("single tenant, single lokistack & es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp, esOut},
					},
				},
				Outputs: []obs.OutputSpec{
					esOutSpec,
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs, esOutSpec)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOut, esOut},
					},
				}
			},
		),
		Entry("multiple tenants, single lokistack & es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp, esOut},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOutAudit, esOut},
					},
					{
						Name:       lokistackPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOutInfra, esOut},
					},
				},
				Outputs: []obs.OutputSpec{
					esOutSpec,
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs, esOutSpec)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOut, esOut},
					},
				}
			},
		),
		Entry("single tenant, multiple lokistack outputs in one pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp, "another-" + lokistackOutApp},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					})
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOut, "another-" + lokistackOut},
					},
				}
			},
		),
		Entry("multiple tenants, multiple lokistack outputs in one pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp, "another-" + lokistackOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOutAudit, "another-" + lokistackOutAudit},
					},
					{
						Name:       lokistackPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOutInfra, "another-" + lokistackOutInfra},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					})
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOut, "another-" + lokistackOut},
					},
				}
			},
		),
		Entry("single tenant, multiple pipelines, multiple lokistacks in each pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp, "another-" + lokistackOutApp},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOutAudit, "bar-" + lokistackOutAudit},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					obs.OutputSpec{
						Name: "foo-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "foo-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					obs.OutputSpec{
						Name: "bar-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "bar-" + lokistackTarget,
								Namespace: "bar-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOut, "another-" + lokistackOut},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOut, "bar-" + lokistackOut},
					},
				}
			},
		),
		Entry("multiple tenants, multiple pipelines, multiple lokistacks in each pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOutApp, "another-" + lokistackOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOutInfra, "another-" + lokistackOutInfra},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{"foo-" + lokistackOutApp, "bar-" + lokistackOutApp},
					},
					{
						Name:       "another-" + lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOutAudit, "bar-" + lokistackOutAudit},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					obs.OutputSpec{
						Name: "foo-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "foo-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					obs.OutputSpec{
						Name: "bar-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "bar-" + lokistackTarget,
								Namespace: "bar-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOut, "another-" + lokistackOut},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOut, "bar-" + lokistackOut},
					},
				}
			},
		),
	)

	DescribeTable("migrate lokistack to otlp outputs/pipelines", func(expSpec obs.ClusterLogForwarderSpec, visit func(spec *obs.ClusterLogForwarderSpec)) {
		clfSpec := initClf(lokistackOtlpOut, true)
		if visit != nil {
			visit(&clfSpec.Spec)
		}

		spec = MigrateLokiStack(clfSpec, utils.NoOptions)
		Expect(spec.Spec).To(Equal(expSpec))
	},
		Entry("single tenant, single lokistack output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOut},
					},
				}
			},
		),
		Entry("multiple tenants, single lokistack output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOtlpOutAudit},
					},
					{
						Name:       lokistackPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOutInfra},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOut},
					},
				}
			},
		),
		Entry("single tenant, single lokistack & es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, esOut},
					},
				},
				Outputs: []obs.OutputSpec{
					esOutSpec,
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs, esOutSpec)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOut, esOut},
					},
				}
			},
		),
		Entry("multiple tenants, single lokistack & es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, esOut},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOtlpOutAudit, esOut},
					},
					{
						Name:       lokistackPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOutInfra, esOut},
					},
				},
				Outputs: []obs.OutputSpec{
					esOutSpec,
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs, esOutSpec)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOut, esOut},
					},
				}
			},
		),
		Entry("single tenant, multiple lokistack outputs in one pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, "another-" + lokistackOtlpOutApp},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					})
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOut, "another-" + lokistackOtlpOut},
					},
				}
			},
		),
		Entry("multiple tenants, multiple lokistack outputs in one pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, "another-" + lokistackOtlpOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{lokistackOtlpOutAudit, "another-" + lokistackOtlpOutAudit},
					},
					{
						Name:       lokistackPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOutInfra, "another-" + lokistackOtlpOutInfra},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					})
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOut, "another-" + lokistackOtlpOut},
					},
				}
			},
		),
		Entry("single tenant, multiple pipelines, multiple lokistacks in each pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, "another-" + lokistackOtlpOutApp},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOtlpOutAudit, "bar-" + lokistackOtlpOutAudit},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
					obs.OutputSpec{
						Name: "foo-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "foo-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
					obs.OutputSpec{
						Name: "bar-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "bar-" + lokistackTarget,
								Namespace: "bar-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
				)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOut, "another-" + lokistackOtlpOut},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOtlpOut, "bar-" + lokistackOtlpOut},
					},
				}
			},
		),
		Entry("multiple tenants, multiple pipelines, multiple lokistacks in each pipeline",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, "another-" + lokistackOtlpOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOutInfra, "another-" + lokistackOtlpOutInfra},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{"foo-" + lokistackOtlpOutApp, "bar-" + lokistackOtlpOutApp},
					},
					{
						Name:       "another-" + lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOtlpOutAudit, "bar-" + lokistackOtlpOutAudit},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
					obs.OutputSpec{
						Name: "foo-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "foo-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
					obs.OutputSpec{
						Name: "bar-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "bar-" + lokistackTarget,
								Namespace: "bar-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
				)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOut, "another-" + lokistackOtlpOut},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOtlpOut, "bar-" + lokistackOtlpOut},
					},
				}
			},
		),
		Entry("multiple tenants, multiple pipelines, multiple lokistacks in each pipeline, only a subset to OTLP out",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{lokistackOtlpOutApp, "another-" + lokistackOutApp},
					},
					{
						Name:       lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOutInfra, "another-" + lokistackOutInfra},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{"foo-" + lokistackOutApp, "bar-" + lokistackOtlpOutApp},
					},
					{
						Name:       "another-" + lokistackPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOutAudit, "bar-" + lokistackOtlpOutAudit},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: "another-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "another-" + lokistackOutInfra,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://another-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/infrastructure",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "bar-" + lokistackOtlpOutAudit,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://bar-test-lokistack-gateway-http.bar-namespace.svc:8080/api/logs/v1/audit" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: "foo-" + lokistackOutAudit,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://foo-test-lokistack-gateway-http.foo-namespace.svc:8080/api/logs/v1/audit",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutApp,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					{
						Name: lokistackOtlpOutInfra,
						Type: obs.OutputTypeOTLP,
						OTLP: &obs.OTLP{
							URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/infrastructure" + lokiOtlpEndpoint,
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Outputs = append(spec.Outputs,
					obs.OutputSpec{
						Name: "another-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "another-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					obs.OutputSpec{
						Name: "foo-" + lokistackOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "foo-" + lokistackTarget,
								Namespace: "foo-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
						},
					},
					obs.OutputSpec{
						Name: "bar-" + lokistackOtlpOut,
						Type: obs.OutputTypeLokiStack,
						LokiStack: &obs.LokiStack{
							Target: obs.LokiStackTarget{
								Name:      "bar-" + lokistackTarget,
								Namespace: "bar-namespace",
							},
							Authentication: &obs.LokiStackAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccount,
								},
							},
							DataModel: obs.LokiStackDataModelOpenTelemetry,
						},
					},
				)
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{lokistackOtlpOut, "another-" + lokistackOut},
					},
					{
						Name:       "another-" + lokistackPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit)},
						OutputRefs: []string{"foo-" + lokistackOut, "bar-" + lokistackOtlpOut},
					},
				}
			},
		),
	)

	DescribeTable(
		"LabelKeys logic for LokiStack tenants", func(labelKeys *obs.LokiStackLabelKeys, tenant string, wantKeys []string) {
			testDefaultKeys := []string{
				"default_one",
				"default_two",
			}
			keys := lokiStackLabelKeysForTenant(labelKeys, tenant, testDefaultKeys)
			Expect(keys).To(Equal(wantKeys))
		},
		Entry(
			"no config",
			nil,
			string(obs.InputTypeApplication),
			nil,
		),
		Entry(
			"empty slices -> still nil",
			&obs.LokiStackLabelKeys{
				Global: []string{},
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{},
				},
			},
			string(obs.InputTypeApplication),
			nil,
		),
		Entry(
			"only global",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"global_one",
				"global_two",
			},
		),
		Entry(
			"only tenant",
			&obs.LokiStackLabelKeys{
				Application: &obs.LokiStackTenantLabelKeys{
					IgnoreGlobal: true,
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"only tenant but with defaults",
			&obs.LokiStackLabelKeys{
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"default_one",
				"default_two",
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"global and tenant",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
				},
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"global_one",
				"global_two",
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"global and tenant, ignore global",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
				},
				Application: &obs.LokiStackTenantLabelKeys{
					IgnoreGlobal: true,
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"tenant_one",
				"tenant_two",
			},
		),
		Entry(
			"global and tenant, ignore duplicates",
			&obs.LokiStackLabelKeys{
				Global: []string{
					"global_one",
					"global_two",
					"common_one",
				},
				Application: &obs.LokiStackTenantLabelKeys{
					LabelKeys: []string{
						"tenant_one",
						"tenant_two",
						"common_one",
					},
				},
			},
			string(obs.InputTypeApplication),
			[]string{
				"common_one",
				"global_one",
				"global_two",
				"tenant_one",
				"tenant_two",
			},
		),
	)
})
