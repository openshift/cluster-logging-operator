package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Validating tech-preview annotation for OTLP output type", func() {
	Context("#ValidateOtlpAnnotation", func() {
		var (
			k8sClient client.Client
			forwarder obs.ClusterLogForwarder
			context   internalcontext.ForwarderContext
			out       obs.OutputSpec
		)

		BeforeEach(func() {
			out = obs.OutputSpec{
				Name: "my-output",
				Type: obs.OutputTypeOTLP,
			}
			forwarder = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{out},
				},
			}
			forwarder.Annotations = map[string]string{"some.other.annotation/for-testing": "true"}
			k8sClient = fake.NewFakeClient()
			context = internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: &forwarder,
			}
		})

		It("should pass validation when type is not OTLP", func() {
			out.Type = obs.OutputTypeHTTP
			forwarder.Spec.Outputs = []obs.OutputSpec{out}
			// Return value is empty when validation passes
			Expect(ValidateOtlpAnnotation(context)).To(BeEmpty())
		})
		It("should pass validation when annotation is included with either value", func() {
			forwarder.Annotations[constants.AnnotationOtlpOutputTechPreview] = "true"
			Expect(ValidateOtlpAnnotation(context)).To(BeEmpty())

			forwarder.Annotations[constants.AnnotationOtlpOutputTechPreview] = "enabled"
			Expect(ValidateOtlpAnnotation(context)).To(BeEmpty())
		})
		It("should fail validation when missing the annotation", func() {
			results := ValidateOtlpAnnotation(context)
			Expect(results).To(ContainElement(ContainSubstring(MissingAnnotationMessage)))
		})
		It("should fail validation when annotation has incorrect value", func() {
			forwarder.Annotations[constants.AnnotationOtlpOutputTechPreview] = "false"
			results := ValidateOtlpAnnotation(context)
			Expect(results).To(ContainElement(ContainSubstring(MissingAnnotationMessage)))
		})
		It("should still fail validation when including other types", func() {
			out2 := obs.OutputSpec{
				Name: "my-out2",
				Type: obs.OutputTypeCloudwatch,
			}
			out3 := obs.OutputSpec{
				Name: "my-out3",
				Type: obs.OutputTypeLoki,
			}
			forwarder.Spec.Outputs = []obs.OutputSpec{out, out2, out3}
			Expect(ValidateOtlpAnnotation(context)).To(ContainElement(ContainSubstring(MissingAnnotationMessage)))
		})
	})
})
