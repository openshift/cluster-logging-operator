package pipelines

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/matchers"
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

	DescribeTable("should fail", func(input, output, filter, reason string) {
		pipelineSpec := initSpec()
		pipelineSpec.InputRefs = append(pipelineSpec.InputRefs, input)
		pipelineSpec.FilterRefs = append(pipelineSpec.FilterRefs, filter)
		pipelineSpec.OutputRefs = append(pipelineSpec.OutputRefs, output)

		cond := validateRef(pipelineSpec, inputMap, outputMap, filterMap)
		Expect(cond).To(matchers.HaveCondition(obs.ValidationCondition, true, reason, `pipeline.* references (input|filter|output) ".*" not found`))
	},
		Entry("when an input does not exist", "missing", "", "", obs.ReasonPipelineInputRefNotFound),
		Entry("when an output does not exist", "", "missing", "", obs.ReasonPipelineOutputRefNotFound),
		Entry("when a filter does not exist", "", "", "missing", obs.ReasonPipelineFilterRefNotFound),
	)

	It("should pass validation when all inputs, outputs and filters exist", func() {
		cond := validateRef(initSpec(), inputMap, outputMap, filterMap)
		Expect(cond).To(BeEmpty())
	})

})
