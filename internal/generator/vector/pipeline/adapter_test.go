package pipeline_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/pipeline"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
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
	)
	Context("#NewPipeline", func() {
		Describe("when adding a prune filter", func() {
			var (
				inputSpecs = []logging.InputSpec{
					{Name: "app-in", Application: &logging.Application{}},
				}
			)

			It("should add prune filter with only defined `in` fields and no `notIn` fields", func() {
				adapter := NewPipeline(0, logging.PipelineSpec{
					Name:       "mypipeline",
					InputRefs:  []string{"app-in"},
					FilterRefs: []string{"my-prune"},
				}, map[string]helpers.InputComponent{
					inputSpecs[0].Name: input.NewInput(inputSpecs[0], "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				}, map[string]*output.Output{},
					map[string]*filter.InternalFilterSpec{
						"my-prune": {
							FilterSpec: &logging.FilterSpec{
								Name: "my-prune",
								Type: logging.FilterPrune,
								FilterTypeSpec: logging.FilterTypeSpec{
									PruneFilterSpec: &logging.PruneFilterSpec{
										In: []string{".foo.test",
											".bar",
											`.foo."@some"."d.f.g.o111-22/333".foo_bar`,
											`.foo.labels."test.dot-with/slashes888"`},
									},
								},
							},
						},
					},
					inputSpecs,
				)
				Expect(adapter.Filters).To(HaveLen(2), "expected a VIAQ and prune filter to be added to the pipeline")
				Expect(mustLoad("adapter_test_prune_inOnly_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
			})
			It("should add prune filter with only defined `notIn` fields and no `in` fields", func() {
				adapter := NewPipeline(0, logging.PipelineSpec{
					Name:       "mypipeline",
					InputRefs:  []string{"app-in"},
					FilterRefs: []string{"my-prune"},
				}, map[string]helpers.InputComponent{
					inputSpecs[0].Name: input.NewInput(inputSpecs[0], "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				}, map[string]*output.Output{},
					map[string]*filter.InternalFilterSpec{
						"my-prune": {
							FilterSpec: &logging.FilterSpec{
								Name: "my-prune",
								Type: logging.FilterPrune,
								FilterTypeSpec: logging.FilterTypeSpec{
									PruneFilterSpec: &logging.PruneFilterSpec{
										NotIn: []string{".kubernetes.labels", ".message", ".foo"},
									},
								},
							},
						},
					},
					inputSpecs,
				)
				Expect(adapter.Filters).To(HaveLen(2), "expected VIA and prune filter to be added to the pipeline")
				Expect(mustLoad("adapter_test_prune_notIn_only_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
			})
			It("should add prune filter with both defined in fields and notIn fields when spec'd", func() {
				adapter := NewPipeline(0, logging.PipelineSpec{
					Name:       "mypipeline",
					InputRefs:  []string{"app-in"},
					FilterRefs: []string{"my-prune"},
				}, map[string]helpers.InputComponent{
					inputSpecs[0].Name: input.NewInput(inputSpecs[0], "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				}, map[string]*output.Output{},
					map[string]*filter.InternalFilterSpec{
						"my-prune": {
							FilterSpec: &logging.FilterSpec{
								Name: "my-prune",
								Type: logging.FilterPrune,
								FilterTypeSpec: logging.FilterTypeSpec{
									PruneFilterSpec: &logging.PruneFilterSpec{
										In:    []string{".kubernetes.labels.foo", ".log_type", ".message"},
										NotIn: []string{".kubernetes.container_name", `.foo.bar."baz/bar"`, `.foo`},
									},
								},
							},
						},
					},
					inputSpecs,
				)
				Expect(adapter.Filters).To(HaveLen(2), "expected a VIA and prune filter to be added to the pipeline")
				Expect(mustLoad("adapter_test_prune_inNotIn_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
			})
		})

		It("should add Kube API Server Audit Policy when spec'd for the pipeline", func() {
			inputSpecs := []logging.InputSpec{
				{Name: "audit-in", Audit: &logging.Audit{
					Sources: []string{logging.AuditSourceKube},
				}},
			}
			adapter := NewPipeline(0, logging.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{inputSpecs[0].Name},
				FilterRefs: []string{"my-audit"},
			}, map[string]helpers.InputComponent{
				inputSpecs[0].Name: input.NewInput(inputSpecs[0], "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
			}, map[string]*output.Output{},
				map[string]*filter.InternalFilterSpec{
					"my-audit": {
						FilterSpec: &logging.FilterSpec{
							Name: "my-audit",
							Type: logging.FilterKubeAPIAudit,
							FilterTypeSpec: logging.FilterTypeSpec{
								KubeAPIAudit: &logging.KubeAPIAudit{
									Rules: []auditv1.PolicyRule{
										{Level: auditv1.LevelRequestResponse, Users: []string{"*apiserver"}}, // Keep full event for user ending in *apiserver
										{Level: auditv1.LevelNone, Verbs: []string{"get"}},                   // Drop other GET requests
										{Level: auditv1.LevelMetadata},                                       // Metadata for everything else.
									},
								},
							},
						},
					},
				},
				inputSpecs,
			)
			Expect(adapter.Filters).To(HaveLen(2), "expected VIAQ and kubeapi filter to be added to the pipeline")
			Expect(mustLoad("adapter_test_kube_api_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
		})

		It("should add drop filter when spec'd for the pipeline", func() {
			inputSpecs := []logging.InputSpec{
				{Name: "app-in", Application: &logging.Application{}},
				{Name: "infra-in", Infrastructure: &logging.Infrastructure{
					Sources: logging.InfrastructureSources.List(),
				}},
			}
			adapter := NewPipeline(0, logging.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{inputSpecs[0].Name, inputSpecs[1].Name},
				FilterRefs: []string{"my-drop-filter"},
			}, map[string]helpers.InputComponent{
				inputSpecs[0].Name: input.NewInput(inputSpecs[0], "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
				inputSpecs[1].Name: input.NewInput(inputSpecs[1], "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
			}, map[string]*output.Output{},
				map[string]*filter.InternalFilterSpec{
					"my-drop-filter": {
						FilterSpec: &logging.FilterSpec{
							Name: "my-drop-filter",
							Type: logging.FilterDrop,
							FilterTypeSpec: logging.FilterTypeSpec{
								DropTestsSpec: &[]logging.DropTest{
									{
										DropConditions: []logging.DropCondition{
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
										DropConditions: []logging.DropCondition{
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
						},
					},
				},
				inputSpecs,
			)
			Expect(adapter.Filters).To(HaveLen(3), "expected VIAQ and drop filter to be added to the pipeline")
			Expect(mustLoad("adapter_test_drop_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
		})
	})
})
