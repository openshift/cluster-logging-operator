package outputs

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	"strings"
)

func Validate(context internalcontext.ForwarderContext) {
	pipelines := internalobs.Pipelines(context.Forwarder.Spec.Pipelines)
	for _, out := range context.Forwarder.Spec.Outputs {
		messages := []string{}
		configs := internalobs.SecretReferencesAsValueReferences(out)
		if out.TLS != nil {
			messages = append(messages, validateURLAccordingToTLS(out)...)
			configs = append(configs, internalobs.ValueReferences(out.TLS.TLSSpec)...)
		}
		messages = append(messages, common.ValidateValueReference(configs, context.Secrets, context.ConfigMaps)...)
		messages = append(messages, validateOutputIsReferencedByPipelines(out, pipelines)...)
		// Validate by output type
		// Note: type 'lokiStack' becomes 'otlp' output type when sending Otel otherwise it becomes 'loki'
		switch out.Type {
		case obs.OutputTypeCloudwatch:
			messages = append(messages, ValidateCloudWatchAuth(out, context)...)
		case obs.OutputTypeHTTP:
			messages = append(messages, validateHttpContentTypeHeaders(out)...)
		case obs.OutputTypeOTLP:
			messages = append(messages, ValidateTechPreviewAnnotation(context)...)
		}
		// Set condition
		if len(messages) > 0 {
			internalobs.SetCondition(&context.Forwarder.Status.OutputConditions,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, out.Name, false, obs.ReasonValidationFailure, strings.Join(messages, ",")))
		} else {
			internalobs.SetCondition(&context.Forwarder.Status.OutputConditions,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, out.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("output %q is valid", out.Name)))
		}
	}
}

func validateOutputIsReferencedByPipelines(output obs.OutputSpec, pipelines internalobs.Pipelines) (results []string) {
	if !pipelines.ReferenceOutput(output) {
		return append(results, "not referenced by any pipeline")
	}
	return results
}
