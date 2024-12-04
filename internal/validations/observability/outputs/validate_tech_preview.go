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
func ValidateTechPreviewAnnotation(context internalcontext.ForwarderContext) (messages []string) {
	enabled := common.IsEnabledAnnotation(context, constants.AnnotationOtlpOutputTechPreview)
	for _, out := range context.Forwarder.Spec.Outputs {
		if out.Type == obsv1.OutputTypeOTLP && !enabled {
			log.V(3).Info("ValidateTechPreviewAnnotation failed", "reason", MissingAnnotationMessage)
			messages = append(messages, fmt.Sprintf("output %q %v", out.Name, MissingAnnotationMessage))
			return messages
		}
	}
	return messages
}
