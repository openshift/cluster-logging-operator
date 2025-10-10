package initialize

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("initOutputs", func() {
	var (
		initContext utils.Options
	)
	BeforeEach(func() {
		initContext = utils.Options{}
	})
	Context("for a LokiStack output", func() {
		var spec obs.ClusterLogForwarder
		BeforeEach(func() {
			spec = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{
						{Type: obs.OutputTypeLokiStack, LokiStack: &obs.LokiStack{}},
					},
				},
			}
		})
		Context("when maxWrite is not spec'd", func() {

			It("should default a value when tuning is not specified", func() {
				result := InitOutputs(spec, initContext)
				Expect(result.Spec.Outputs[0].LokiStack.Tuning.MaxWrite.String()).To(Equal(DefaultMaxWriteLokiStack))
			})
			It("should default a value when the tuning.maxWrite is not specified", func() {
				spec.Spec.Outputs[0].LokiStack.Tuning = &obs.LokiTuningSpec{}
				result := InitOutputs(spec, initContext)
				Expect(result.Spec.Outputs[0].LokiStack.Tuning.MaxWrite.String()).To(Equal(DefaultMaxWriteLokiStack))
			})
		})
		It("should honor the value spec'd by tuning.maxWrite", func() {
			maxWrite, _ := resource.ParseQuantity("1024Mi")
			spec.Spec.Outputs[0].LokiStack.Tuning = &obs.LokiTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					MaxWrite: &maxWrite,
				},
			}
			result := InitOutputs(spec, initContext)
			Expect(result.Spec.Outputs[0].LokiStack.Tuning.MaxWrite.String()).To(Equal(maxWrite.String()))
		})
	})
})
