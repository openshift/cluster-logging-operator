package initialize

import (
	"github.com/openshift/cluster-logging-operator/internal/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("MigrateLokiStack", func() {
	const (
		esOut             = "es-out"
		lokistackOut      = "lokistack-out"
		lokistackTarget   = "test-lokistack"
		saName            = "test-sa"
		lokistackPipeline = "test-lokistack-pipeline"
		lokistackOutApp   = lokistackOut + "-" + string(obs.InputTypeApplication)
		lokistackOutAudit = lokistackOut + "-" + string(obs.InputTypeAudit)
		lokistackOutInfra = lokistackOut + "-" + string(obs.InputTypeInfrastructure)
	)

	var (
		spec    obs.ClusterLogForwarder
		initClf = func() obs.ClusterLogForwarder {
			return obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{
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
										From: obs.BearerTokenFromServiceAccountToken,
									},
								},
							},
						},
					},
				},
			}
		}
	)

	DescribeTable("migrate lokistack to loki outputs/pipelines", func(expSpec obs.ClusterLogForwarderSpec, visit func(spec *obs.ClusterLogForwarderSpec)) {
		clfSpec := initClf()
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccountToken,
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
					{
						Name: lokistackOutApp,
						Type: obs.OutputTypeLoki,
						Loki: &obs.Loki{
							URLSpec: obs.URLSpec{
								URL: "https://test-lokistack-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
							},
							Authentication: &obs.HTTPAuthentication{
								Token: &obs.BearerToken{
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
		Entry("mulitple tenants, multiple pipelines, multiple lokistacks in each pipeline",
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
									From: obs.BearerTokenFromServiceAccountToken,
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
})
