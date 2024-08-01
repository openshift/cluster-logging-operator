package pipeline_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/pipeline"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

type FakeInputAdapter struct {
	ids []string
}

func (m FakeInputAdapter) InputIDs() []string {
	return m.ids
}

var _ = Describe("pipeline/adapter.go", func() {
	var (
		mustLoad = func(expFile string) string {

			exp, err := tomlContent.ReadFile(expFile)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", exp, err))
			}
			return string(exp)
		}
		secrets = map[string]*corev1.Secret{}
	)
	Context("#NewPipeline", func() {
		Describe("when adding a prune filter", func() {
			var (
				inputSpecs = []obs.InputSpec{
					{
						Name:        "app-in",
						Type:        obs.InputTypeApplication,
						Application: &obs.Application{},
					},
				}
			)

			It("should add prune filter with only defined `in` fields and no `notIn` fields", func() {
				adapter := NewPipeline(0, obs.PipelineSpec{
					Name:       "mypipeline",
					InputRefs:  []string{"app-in"},
					FilterRefs: []string{"my-prune"},
				}, map[string]helpers.InputComponent{
					inputSpecs[0].Name: input.NewInput(inputSpecs[0], secrets, "", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				}, map[string]*output.Output{},
					filter.NewInternalFilterMap(map[string]*obs.FilterSpec{
						"my-prune": {
							Name: "my-prune",
							Type: obs.FilterTypePrune,
							PruneFilterSpec: &obs.PruneFilterSpec{
								In: []obs.FieldPath{".foo.test",
									".bar",
									`.foo."@some"."d.f.g.o111-22/333".foo_bar`,
									`.foo.labels."test.dot-with/slashes888"`},
							},
						},
					}),
					inputSpecs,
				)
				Expect(adapter.Filters).To(HaveLen(3), "expected a viaq, prune and dedot filter to be added to the pipeline")
				Expect(mustLoad("adapter_test_prune_inOnly_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
			})
			It("should add prune filter with only defined `notIn` fields and no `in` fields", func() {
				adapter := NewPipeline(0, obs.PipelineSpec{
					Name:       "mypipeline",
					InputRefs:  []string{"app-in"},
					FilterRefs: []string{"my-prune"},
				}, map[string]helpers.InputComponent{
					inputSpecs[0].Name: input.NewInput(inputSpecs[0], secrets, "", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				}, map[string]*output.Output{},
					filter.NewInternalFilterMap(map[string]*obs.FilterSpec{
						"my-prune": {
							Name: "my-prune",
							Type: obs.FilterTypePrune,
							PruneFilterSpec: &obs.PruneFilterSpec{
								NotIn: []obs.FieldPath{".kubernetes.labels", ".message", ".foo"},
							},
						},
					}),
					inputSpecs,
				)
				Expect(adapter.Filters).To(HaveLen(3), "expected viaq, prune and dedot filters to be added to the pipeline")
				Expect(mustLoad("adapter_test_prune_notIn_only_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
			})
			It("should add prune filter with both defined in fields and notIn fields when spec'd", func() {
				adapter := NewPipeline(0, obs.PipelineSpec{
					Name:       "mypipeline",
					InputRefs:  []string{"app-in"},
					FilterRefs: []string{"my-prune"},
				}, map[string]helpers.InputComponent{
					inputSpecs[0].Name: input.NewInput(inputSpecs[0], secrets, "", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				}, map[string]*output.Output{},
					filter.NewInternalFilterMap(map[string]*obs.FilterSpec{
						"my-prune": {
							Name: "my-prune",
							Type: obs.FilterTypePrune,
							PruneFilterSpec: &obs.PruneFilterSpec{
								In:    []obs.FieldPath{".kubernetes.labels.foo", ".log_type", ".message"},
								NotIn: []obs.FieldPath{".kubernetes.container_name", `.foo.bar."baz/bar"`, `.foo`},
							},
						},
					}),
					inputSpecs,
				)
				Expect(adapter.Filters).To(HaveLen(3), "expected a viaq, prune and dedot filter to be added to the pipeline")
				Expect(mustLoad("adapter_test_prune_inNotIn_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
			})
		})

		It("should add Kube API Server Audit Policy when spec'd for the pipeline", func() {
			inputSpecs := []obs.InputSpec{
				{Name: "audit-in", Type: obs.InputTypeAudit, Audit: &obs.Audit{
					Sources: []obs.AuditSource{obs.AuditSourceKube},
				}},
			}
			adapter := NewPipeline(0, obs.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{inputSpecs[0].Name},
				FilterRefs: []string{"my-audit"},
			}, map[string]helpers.InputComponent{
				inputSpecs[0].Name: input.NewInput(inputSpecs[0], secrets, "", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
			}, map[string]*output.Output{},
				filter.NewInternalFilterMap(map[string]*obs.FilterSpec{
					"my-audit": {
						Name: "my-audit",
						Type: obs.FilterTypeKubeAPIAudit,
						KubeAPIAudit: &obs.KubeAPIAudit{
							Rules: []auditv1.PolicyRule{
								{Level: auditv1.LevelRequestResponse, Users: []string{"*apiserver"}}, // Keep full event for user ending in *apiserver
								{Level: auditv1.LevelNone, Verbs: []string{"get"}},                   // Drop other GET requests
								{Level: auditv1.LevelMetadata},                                       // Metadata for everything else.
							},
						},
					},
				}),
				inputSpecs,
			)
			Expect(adapter.Filters).To(HaveLen(3), "expected viaq, kubeapi and dedot filters to be added to the pipeline")
			Expect(mustLoad("adapter_test_kube_api_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
		})

		It("should add drop filter when spec'd for the pipeline", func() {
			inputSpecs := []obs.InputSpec{
				{Name: "app-in", Type: obs.InputTypeApplication, Application: &obs.Application{}},
				{Name: "infra-in", Type: obs.InputTypeInfrastructure, Infrastructure: &obs.Infrastructure{
					Sources: obs.InfrastructureSources,
				}},
			}
			adapter := NewPipeline(0, obs.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{inputSpecs[0].Name, inputSpecs[1].Name},
				FilterRefs: []string{"my-drop-filter"},
			}, map[string]helpers.InputComponent{
				inputSpecs[0].Name: input.NewInput(inputSpecs[0], secrets, "", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				inputSpecs[1].Name: input.NewInput(inputSpecs[1], secrets, "", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
			}, map[string]*output.Output{},
				filter.NewInternalFilterMap(map[string]*obs.FilterSpec{
					"my-drop-filter": {
						Name: "my-drop-filter",
						Type: obs.FilterTypeDrop,
						DropTestsSpec: []obs.DropTest{
							{
								DropConditions: []obs.DropCondition{
									{
										Field:      ".kubernetes.namespace_name",
										NotMatches: "very-important",
									},
									{
										Field:   ".level",
										Matches: "warning|error|critical",
									},
								},
							},
							{
								DropConditions: []obs.DropCondition{
									{
										Field:   ".message",
										Matches: "foobar",
									},
									{
										Field:      `.kubernetes.namespace_labels."test-dashes/slashes"`,
										NotMatches: "true",
									},
								},
							},
						},
					},
				}),
				inputSpecs,
			)
			Expect(adapter.Filters).To(HaveLen(4), "expected journal, viaq, drop and dedot filters to be added to the pipeline")
			Expect(mustLoad("adapter_test_drop_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
		})
	})
})
