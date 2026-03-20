package adapters_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type FakeElement struct {
	UpdatedFromAddPostFilter string
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
		internalFilterMap = map[string]*adapters.InternalFilterSpec{
			"fakeFilter": {
				Factory: func(inputs ...string) types.Transform {
					condition := fmt.Sprintf("updatedFromAddPostfilter: %v", fakeElement.UpdatedFromAddPostFilter)
					return transforms.NewRemap(condition, inputs...)
				},
			},
			"dropFilter": {
				FilterSpec: &obs.FilterSpec{
					Name: "dropFilter",
					Type: obs.FilterTypeDrop,
				},
				Factory: func(inputs ...string) types.Transform {
					condition := "fakeElementVRL"
					return transforms.NewRemap(condition, inputs...)
				},
			},
		}
	)
	BeforeEach(func() {
		i := adapters.NewInput(inputSpecs[0])
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
		It("should initialize the pipeline adapter and wire them to the output adapters", func() {
			adapter := adapters.NewPipeline(0, obs.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{"app-in"},
				FilterRefs: []string{"dropFilter"},
				OutputRefs: []string{"referenced"},
			}, inputMap,
				outputMap,
				internalFilterMap,
				inputSpecs,
				func(p *adapters.Pipeline) {
					p.FilterRefs = append(p.FilterRefs, "fakeFilter")
					fakeElement.UpdatedFromAddPostFilter = p.Name()
				},
			)
			Expect(adapter.Filters).To(HaveLen(2), "expected the filter and post-filter to be added to the pipeline")
			Expect(api.Transforms{
				"pipeline_mypipeline_dropfilter_0": transforms.NewRemap("fakeElementVRL", "input_app_in_container_meta"),
				"pipeline_mypipeline_fakefilter_1": transforms.NewRemap("updatedFromAddPostfilter: mypipeline", "pipeline_mypipeline_dropfilter_0"),
			}).To(Equal(adapter.Transforms()))
			Expect(outputMap["referenced"].Inputs()).To(Equal([]string{"pipeline_mypipeline_fakefilter_1"}))
			Expect(outputMap["notReferenced"].Inputs()).To(BeNil(), "Exp. the unreferenced output to not have the filter as an input")
		})
	})

	Describe("#NewPipelineFilter", func() {

		It("should add drop filter when spec'd for the pipeline", func() {
			adapter := adapters.NewPipeline(0, obs.PipelineSpec{
				Name:       "mypipeline",
				InputRefs:  []string{"app-in"},
				FilterRefs: []string{"dropFilter"},
				OutputRefs: []string{"referenced"},
			}, inputMap,
				outputMap,
				internalFilterMap,
				inputSpecs,
				func(p *adapters.Pipeline) {},
			)
			Expect(adapter.Filters).To(HaveLen(1), "")
			Expect(api.Transforms{
				"pipeline_mypipeline_dropfilter_0": transforms.NewRemap("fakeElementVRL", "input_app_in_container_meta"),
			}).To(Equal(adapter.Transforms()))
			Expect(outputMap["referenced"].Inputs()).To(Equal([]string{"pipeline_mypipeline_dropfilter_0"}))
			Expect(outputMap["notReferenced"].Inputs()).To(BeNil(), "Exp. the unreferenced output to not have the filter as an input")

		})
	})
})
