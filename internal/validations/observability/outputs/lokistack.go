package outputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

// ValidateLokistackOTLPForAnnotation validates lokistacks designated to ingest OTLP data has the OTLP tech preview annotation
func ValidateLokistackOTLPForAnnotation(context internalcontext.ForwarderContext) {
	clf := context.Forwarder

	cond := internalobs.NewCondition(obs.ConditionTypeValidLokistackOTLPOutputs, obs.ConditionTrue, obs.ReasonValidationSuccess, "")
	for _, outSpec := range clf.Spec.Outputs {
		if outSpec.Type == obs.OutputTypeLokiStack &&
			(outSpec.LokiStack != nil && outSpec.LokiStack.DataModel == obs.LokiStackDataModelOpenTelemetry) {
			// Check if OTLP tech preview annotation is defined
			if _, ok := clf.Annotations[constants.AnnotationOtlpOutputTechPreview]; !ok {
				cond.Status = obs.ConditionFalse
				cond.Reason = obs.ReasonValidationFailure
				cond.Message = "missing tech-preview annotation for OTLP output"
				internalobs.SetCondition(&context.Forwarder.Status.Conditions, cond)
				return
			}
		}
	}
	internalobs.SetCondition(&context.Forwarder.Status.Conditions, cond)
}
