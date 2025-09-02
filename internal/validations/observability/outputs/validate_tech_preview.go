package outputs

import (
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
)

const MissingAnnotationMessage = "requires a valid tech-preview annotation"

// ValidateTechPreviewAnnotation verifies the tech-preview annotation for outputs sending OTEL data
func ValidateTechPreviewAnnotation(out obsv1.OutputSpec, context internalcontext.ForwarderContext) (messages []string) {
	enabled := common.IsEnabledAnnotation(*context.Forwarder, constants.AnnotationOtlpOutputTechPreview)
	if out.Type == obsv1.OutputTypeOTLP && !enabled {
		log.V(3).Info("ValidateTechPreviewAnnotation failed", "reason", MissingAnnotationMessage)
		messages = append(messages, fmt.Sprintf("output %q %v", out.Name, MissingAnnotationMessage))
	} else if out.Type == obsv1.OutputTypeLokiStack && out.LokiStack != nil && out.LokiStack.DataModel == obsv1.LokiStackDataModelOpenTelemetry && !enabled {
		log.V(3).Info("ValidateTechPreviewAnnotation failed", "reason", MissingAnnotationMessage)
		messages = append(messages, fmt.Sprintf("output %q of type, %q, with dataModel, %q, %v", out.Name, obsv1.OutputTypeLokiStack, obsv1.LokiStackDataModelOpenTelemetry, MissingAnnotationMessage))
	}
	return messages
}
