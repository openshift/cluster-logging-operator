package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("DropUnreferencedOutputs", func() {

	var (
		spec loggingv1.ClusterLogForwarderSpec
	)

	BeforeEach(func() {
		spec = loggingv1.ClusterLogForwarderSpec{
			Outputs: []loggingv1.OutputSpec{
				{Name: "dropme"},
				{Name: "foo"},
			},
			Pipelines: []loggingv1.PipelineSpec{
				{OutputRefs: []string{"foo"}},
			},
		}
	})

	It("should drop outputs that are not referenced by any pipeline", func() {
		result, _, conditions := DropUnreferencedOutputs("", "", spec, nil, nil, "", "")
		Expect(result).To(Equal(loggingv1.ClusterLogForwarderSpec{
			Outputs: []loggingv1.OutputSpec{
				{Name: "foo"},
			},
			Pipelines: []loggingv1.PipelineSpec{
				{OutputRefs: []string{"foo"}},
			},
		}))
		Expect(conditions).To(HaveCondition(OutputDroppedCondition, true, OutputNotReferencedReason, ""))
	})

	It("should ignore outputs that are referenced by a pipeline", func() {
		spec.Pipelines = append(spec.Pipelines, loggingv1.PipelineSpec{
			OutputRefs: []string{"dropme"},
		})
		result, _, conditions := DropUnreferencedOutputs("", "", spec, nil, nil, "", "")
		Expect(result).To(Equal(loggingv1.ClusterLogForwarderSpec{
			Outputs: []loggingv1.OutputSpec{
				{Name: "dropme"},
				{Name: "foo"},
			},
			Pipelines: []loggingv1.PipelineSpec{
				{OutputRefs: []string{"foo"}},
				{OutputRefs: []string{"dropme"}},
			},
		}))
		Expect(conditions).To(BeEmpty())
	})
})
