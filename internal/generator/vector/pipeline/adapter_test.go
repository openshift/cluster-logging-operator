package pipeline_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/pipeline"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

type FakeElement struct {
	ID                       string
	Inputs                   []string
	UpdatedFromAddPostFilter string
}

func (f *FakeElement) Name() string {
	return "fakeElement"
}

func (f *FakeElement) Template() string {
	return `{{define "fakeElement"}}
{{.ID}}
inputs: {{.Inputs}}
updatedFromAddPostfilter: {{.UpdatedFromAddPostFilter}}
{{end}}
`
}
func (f *FakeElement) VRL() (string, error) {
	return "fakeElementVRL", nil
}

var _ = Describe("Pipeline adapters", func() {

	var (
		inputSpecs = []obs.InputSpec{
			{
				Name:        "app-in",
				Type:        obs.InputTypeApplication,
				Application: &obs.Application{},
			},
		}
		inputMap map[string]helpers.InputComponent

		outputMap         map[string]*adapters.Output
		fakeElement       = &FakeElement{}
		internalFilterMap = map[string]*filter.InternalFilterSpec{
			"fakeFilter": {
				SuppliesTransform: true,
				TranformFactory: func(id string, inputs ...string) framework.Element {
					fakeElement.ID = id
					fakeElement.Inputs = inputs
					return fakeElement
				},
			},
			"dropFilter": {
				FilterSpec: &obs.FilterSpec{
					Name: "dropFilter",
					Type: obs.FilterTypeDrop,
				},
				RemapFilter: fakeElement,
			},
		}
	)
	BeforeEach(func() {
		i := observability.NewInput(inputSpecs[0])
		i.Ids = []string{"input_app_in_container_meta"}
		inputMap = map[string]helpers.InputComponent{
			inputSpecs[0].Name: i,
		}
		outputMap = map[string]*adapters.Output{
			"referenced":    {},
			"notReferenced": {},
		}

	})

	Describe("#NewPipeline", func() {
		It("should initialize the pipeline and wire them to the output adapters", func() {
			adapter := NewPipeline(0, obs.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{"app-in"},
				FilterRefs: []string{"fakeFilter"},
				OutputRefs: []string{"referenced"},
			}, inputMap,
				outputMap,
				internalFilterMap,
				inputSpecs,
				func(p *Pipeline) {
					fakeElement.UpdatedFromAddPostFilter = p.Name()
				},
			)
			Expect(adapter.Filters).To(HaveLen(1), "expected the filter and post-filter to be added to the pipeline")
			Expect(`
pipeline_mypipeline_fakefilter_0
inputs: [input_app_in_container_meta]
updatedFromAddPostfilter: mypipeline
`).To(EqualConfigFrom(adapter.Elements()))
			Expect(outputMap["referenced"].Inputs()).To(Equal([]string{"pipeline_mypipeline_fakefilter_0"}))
			Expect(outputMap["notReferenced"].Inputs()).To(BeNil(), "Exp. the unreferenced output to not have the filter as an input")
		})
	})

	Describe("#NewPipelineFilter", func() {

		It("should add drop filter when spec'd for the pipeline", func() {
			adapter := NewPipeline(0, obs.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{"app-in"},
				FilterRefs: []string{"dropFilter"},
				OutputRefs: []string{"referenced"},
			}, inputMap,
				outputMap,
				internalFilterMap,
				inputSpecs,
				func(p *Pipeline) {},
			)
			Expect(adapter.Filters).To(HaveLen(1), "")
			Expect(`
[transforms.pipeline_mypipeline_dropfilter_0]
type = "filter"
inputs = ["input_app_in_container_meta"]
condition = '''
fakeElementVRL
'''
`).To(EqualConfigFrom(adapter.Elements()))
			Expect(outputMap["referenced"].Inputs()).To(Equal([]string{"pipeline_mypipeline_dropfilter_0"}))
			Expect(outputMap["notReferenced"].Inputs()).To(BeNil(), "Exp. the unreferenced output to not have the filter as an input")

		})
	})
})
