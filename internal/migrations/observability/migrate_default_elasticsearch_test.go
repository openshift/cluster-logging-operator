package observability

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("MigrateDefaultElasticsearch", func() {
	const (
		esOut             = "es-out"
		defaultESOut      = api.DefaultEsName
		saName            = "test-sa"
		defaultESPipeline = "test-default-pipeline"
		defaultEsOutApp   = defaultESOut + "-" + string(obs.InputTypeApplication)
		defaultEsOutAudit = defaultESOut + "-" + string(obs.InputTypeAudit)
		defaultEsOutInfra = defaultESOut + "-" + string(obs.InputTypeInfrastructure)
	)

	var (
		spec  obs.ClusterLogForwarderSpec
		esTls = &obs.OutputTLSSpec{
			TLSSpec: obs.TLSSpec{
				CA: &obs.ConfigReference{
					Secret: &v1.LocalObjectReference{
						Name: constants.CollectorName,
					},
					Key: constants.TrustedCABundleKey,
				},
				Certificate: &obs.ConfigReference{
					Secret: &v1.LocalObjectReference{
						Name: constants.CollectorName,
					},
					Key: constants.ClientCertKey,
				},
				Key: &obs.SecretConfigReference{
					Secret: &v1.LocalObjectReference{
						Name: constants.CollectorName,
					},
					Key: constants.ClientPrivateKey,
				},
			},
		}
		initClf = func() obs.ClusterLogForwarderSpec {
			return obs.ClusterLogForwarderSpec{
				Outputs: []obs.OutputSpec{
					{
						Name: defaultESOut,
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
						TLS: esTls,
					},
				},
			}
		}
	)

	DescribeTable("migrate defaultEs to outputs/pipelines", func(expSpec obs.ClusterLogForwarderSpec, visit func(spec *obs.ClusterLogForwarderSpec)) {
		clfSpec := initClf()
		if visit != nil {
			visit(&clfSpec)
		}

		spec, _ = MigrateDefaultElasticsearch(clfSpec)
		Expect(spec).To(Equal(expSpec))
	},
		Entry("single tenant, default es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{defaultEsOutApp},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: defaultEsOutApp,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: AppIndex,
							},
						},
						TLS: esTls,
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{defaultESOut},
					},
				}
			},
		),
		Entry("multiple tenants, default es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{defaultEsOutApp},
					},
					{
						Name:       defaultESPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{defaultEsOutAudit},
					},
					{
						Name:       defaultESPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{defaultEsOutInfra},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: defaultEsOutApp,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: AppIndex,
							},
						},
						TLS: esTls,
					},
					{
						Name: defaultEsOutAudit,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: AuditIndex,
							},
						},
						TLS: esTls,
					},
					{
						Name: defaultEsOutInfra,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: InfraIndex,
							},
						},
						TLS: esTls,
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{defaultESOut},
					},
				}
			},
		),
		Entry("single tenant, default-es & es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{defaultEsOutApp, esOut},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: defaultEsOutApp,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: AppIndex,
							},
						},
						TLS: esTls,
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{defaultESOut, esOut},
					},
				}
			},
		),
		Entry("multiple tenants, default es & es output",
			obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication)},
						OutputRefs: []string{defaultEsOutApp, esOut},
					},
					{
						Name:       defaultESPipeline + "-1",
						InputRefs:  []string{string(obs.InputTypeAudit)},
						OutputRefs: []string{defaultEsOutAudit, esOut},
					},
					{
						Name:       defaultESPipeline + "-2",
						InputRefs:  []string{string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{defaultEsOutInfra, esOut},
					},
				},
				Outputs: []obs.OutputSpec{
					{
						Name: defaultEsOutApp,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: AppIndex,
							},
						},
						TLS: esTls,
					},
					{
						Name: defaultEsOutAudit,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: AuditIndex,
							},
						},
						TLS: esTls,
					},
					{
						Name: defaultEsOutInfra,
						Type: obs.OutputTypeElasticsearch,
						Elasticsearch: &obs.Elasticsearch{
							URLSpec: obs.URLSpec{
								URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
							},
							Version: 6,
							IndexSpec: obs.IndexSpec{
								Index: InfraIndex,
							},
						},
						TLS: esTls,
					},
				},
			},
			func(spec *obs.ClusterLogForwarderSpec) {
				spec.Pipelines = []obs.PipelineSpec{
					{
						Name:       defaultESPipeline,
						InputRefs:  []string{string(obs.InputTypeApplication), string(obs.InputTypeAudit), string(obs.InputTypeInfrastructure)},
						OutputRefs: []string{defaultESOut, esOut},
					},
				}
			},
		),
	)
})
