package pipeline_test

import (
	"fmt"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
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
		It("should add Kube API Server Audit Policy when spec'd for the pipeline", func() {
			adapter := NewPipeline(0, logging.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{"audit-in"},
				FilterRefs: []string{"my-audit"},
			}, map[string]helpers.InputComponent{
				"audit-in": input.NewInput(logging.InputSpec{Name: "audit-in", Application: &logging.Application{}}, "", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, nil),
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
			)
			Expect(adapter.Filters).To(HaveLen(1), "expected a filter to be added to the pipeline")
			Expect(mustLoad("adapter_test_kube_api_filter.toml")).To(EqualConfigFrom(adapter.Elements()))
		})

		It("should configure all inputRefs to all the outputRefs", func() {

			outputAdapter := output.NewOutput(logging.OutputSpec{
				Name: "mylogstore",
			}, nil, nil)
			otherOutputAdapter := output.NewOutput(logging.OutputSpec{
				Name: "other",
			}, nil, nil)
			NewPipeline(0, logging.PipelineSpec{
				Name: "mypipeline",
				InputRefs: []string{
					logging.InputNameApplication,
					logging.InputNameInfrastructure,
					logging.InputNameAudit,
				},
				OutputRefs: []string{"mylogstore", "other"},
				FilterRefs: []string{},
			}, map[string]helpers.InputComponent{
				logging.InputNameApplication:    FakeInputAdapter{ids: []string{logging.InputNameApplication}},
				logging.InputNameInfrastructure: FakeInputAdapter{ids: []string{logging.InputNameInfrastructure}},
				logging.InputNameAudit:          FakeInputAdapter{ids: []string{logging.InputNameAudit}},
			}, map[string]*output.Output{
				"mylogstore": outputAdapter,
				"other":      otherOutputAdapter,
			},
				map[string]*filter.InternalFilterSpec{},
			)
			inputs := outputAdapter.Inputs()
			sort.Strings(inputs)
			Expect(inputs).To(Equal([]string{logging.InputNameApplication, logging.InputNameAudit, logging.InputNameInfrastructure}))

			inputs = otherOutputAdapter.Inputs()
			sort.Strings(inputs)
			Expect(inputs).To(Equal([]string{logging.InputNameApplication, logging.InputNameAudit, logging.InputNameInfrastructure}))
		})
	})
})
