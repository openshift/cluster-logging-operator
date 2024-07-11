package outputs

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	"strings"
)

func Validate(context internalcontext.ForwarderContext) {

	for _, out := range context.Forwarder.Spec.Outputs {
		configs := internalobs.SecretKeysAsConfigMapOrSecretKeys(out)
		if out.TLS != nil {
			configs = append(configs, internalobs.ConfigReferences(out.TLS.TLSSpec)...)
		}
		messages := common.ValidateConfigReferences(configs, context.Secrets, context.ConfigMaps)
		if out.Type == obsv1.OutputTypeCloudwatch {
			messages = append(messages, ValidateCloudWatchAuth(out)...)
		}
		if len(messages) > 0 {
			internalobs.SetCondition(&context.Forwarder.Status.Outputs,
				internalobs.NewConditionFromPrefix(obsv1.ConditionTypeValidOutputPrefix, out.Name, false, obsv1.ReasonValidationFailure, strings.Join(messages, ",")))
		} else {
			internalobs.SetCondition(&context.Forwarder.Status.Outputs,
				internalobs.NewConditionFromPrefix(obsv1.ConditionTypeValidOutputPrefix, out.Name, true, obsv1.ReasonValidationSuccess, fmt.Sprintf("output %q is valid", out.Name)))
		}
	}
}
