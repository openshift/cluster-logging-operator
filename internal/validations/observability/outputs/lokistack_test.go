package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validate lokistacks designated to ingest OTLP data", func() {
	var (
		forwarder *obs.ClusterLogForwarder
		context   *internalcontext.ForwarderContext
		out       *obs.OutputSpec
	)

	BeforeEach(func() {
		out = &obs.OutputSpec{
			Name:      "my-lokistack-otlp",
			Type:      obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{},
		}
		forwarder = &obs.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{constants.AnnotationOtlpOutputTechPreview: "enabled"},
			},
			Spec: obs.ClusterLogForwarderSpec{},
		}

	})
	It("should pass validation if lokistacks are designated for OTLP output and has the OTLP tech preview annotation", func() {
		out.LokiStack = &obs.LokiStack{
			DataModel: obs.LokiStackDataModelOpenTelemetry,
		}
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *out)

		context = &internalcontext.ForwarderContext{
			Forwarder: forwarder,
		}
		ValidateLokistackOTLPForAnnotation(*context)
		Expect(forwarder.Status.Conditions).To(HaveCondition(obs.ConditionTypeValidLokistackOTLPOutputs, true, obs.ReasonValidationSuccess, ""))
	})

	It("should fail validation if lokistacks are designated for OTLP output but OTLP tech preview annotation is missing", func() {
		out.LokiStack = &obs.LokiStack{
			DataModel: obs.LokiStackDataModelOpenTelemetry,
		}
		forwarder.Spec.Outputs = append(forwarder.Spec.Outputs, *out)
		forwarder.Annotations = map[string]string{}

		context = &internalcontext.ForwarderContext{
			Forwarder: forwarder,
		}
		ValidateLokistackOTLPForAnnotation(*context)
		Expect(forwarder.Status.Conditions).To(HaveCondition(obs.ConditionTypeValidLokistackOTLPOutputs, false, obs.ReasonValidationFailure, "missing tech-preview annotation for OTLP output"))
	})
})
