package outputs

import (
	"fmt"
	"github.com/golang-collections/collections/set"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func Validate(context internalcontext.ForwarderContext) {
	secrets := helpers.Secrets(map[string]*corev1.Secret{})
	roleARNs := set.New()
	for _, out := range context.Forwarder.Spec.Outputs {
		configs := internalobs.SecretReferencesAsValueReferences(out)
		if out.TLS != nil {
			configs = append(configs, internalobs.ValueReferences(out.TLS.TLSSpec)...)
		}
		messages := common.ValidateValueReference(configs, context.Secrets, context.ConfigMaps)
		// Validate by output type
		switch out.Type {
		case obs.OutputTypeCloudwatch:
			if out.Cloudwatch.Authentication.Type == obs.CloudwatchAuthTypeIAMRole {
				roleARNs.Insert(secrets.AsString(out.Cloudwatch.Authentication.IAMRole.RoleARN))
			}
		case obs.OutputTypeOTLP:
			messages = append(messages, ValidateOtlpAnnotation(context)...)
		}

		if roleARNs.Len() > 1 {
			messages = append(messages, "found various CloudWatch RoleARN auth in outputs spec")
		}
		// Set condition
		if len(messages) > 0 {
			internalobs.SetCondition(&context.Forwarder.Status.Outputs,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, out.Name, false, obs.ReasonValidationFailure, strings.Join(messages, ",")))
		} else {
			internalobs.SetCondition(&context.Forwarder.Status.Outputs,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, out.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("output %q is valid", out.Name)))
		}
	}
}
