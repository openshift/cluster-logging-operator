package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("Pipeline validation #validateRef", func() {

	var (
		initSpec = func() obs.PipelineSpec {
			return obs.PipelineSpec{
				Name:       "myPipeline",
				InputRefs:  []string{"anInput"},
				OutputRefs: []string{"anOutput"},
				FilterRefs: []string{"aFilter"},
			}
		}
		inputMap = map[string]obs.InputSpec{
			"anInput": {},
		}
		outputMap = map[string]obs.OutputSpec{
			"anOutput": {},
		}
		filterMap = map[string]*obs.FilterSpec{
			"aFilter": {},
		}
	)

	DescribeTable("should fail", func(input, output, filter, messageRE string) {
		pipelineSpec := initSpec()
		pipelineSpec.InputRefs = append(pipelineSpec.InputRefs, input)
		pipelineSpec.FilterRefs = append(pipelineSpec.FilterRefs, filter)
		pipelineSpec.OutputRefs = append(pipelineSpec.OutputRefs, output)

		cond := validateRef(pipelineSpec, inputMap, outputMap, filterMap)
		Expect(cond).To(ContainElement(MatchRegexp(messageRE)))
	},
		Entry("when an input does not exist", "missing", "", "", `^inputs\[.*\]$`),
		Entry("when an output does not exist", "", "missing", "", `outputs\[.*\]`),
		Entry("when a filter does not exist", "", "", "missing", `filters\[.*\]`),
	)

	It("should pass validation when all inputs, outputs and filters exist", func() {
		cond := validateRef(initSpec(), inputMap, outputMap, filterMap)
		Expect(cond).To(BeEmpty())
	})

})
