package outputs

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"strings"
)

const MissingAnnotationMessage = "missing tech-preview annotation"

// ValidateOtlpAnnotation verifies the tech-preview annotation for OTLP output
func ValidateOtlpAnnotation(context internalcontext.ForwarderContext) (results []string) {
	for _, out := range context.Forwarder.Spec.Outputs {
		if out.Type == obsv1.OutputTypeOTLP {
			if value, ok := context.Forwarder.Annotations[constants.AnnotationOtlpOutputTechPreview]; ok {
				if strings.ToLower(value) == "true" || strings.ToLower(value) == "enabled" {
					// good valid response is nil slice
					return results
				}
			}
			// annotation missing or invalid value
			return []string{fmt.Sprintf("output type %q is %v", string(out.Type), MissingAnnotationMessage)}
		}
	}
	// output type not found so return nil
	return results
}
