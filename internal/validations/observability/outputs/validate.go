package outputs

import (
	"fmt"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Validate(context internalcontext.ForwarderContext) (_ common.AttributeConditionType, results []metav1.Condition) {

	for _, out := range context.Forwarder.Spec.Outputs {
		configs := internalobs.SecretKeysAsConfigMapOrSecretKeys(out)
		if out.TLS != nil {
			configs = append(configs, internalobs.ConfigMapOrSecretKeys(out.TLS.TLSSpec)...)
		}
		messages := common.ValidateConfigMapOrSecretKey(configs, context.Secrets, context.ConfigMaps)
		if out.Type == obsv1.OutputTypeCloudwatch {
			messages = append(messages, ValidateCloudWatchAuth(out)...)
		}
		if len(messages) > 0 {
			results = append(results,
				internalobs.NewConditionFromPrefix(obsv1.ConditionValidOutputPrefix, out.Name, false, obsv1.ReasonValidationFailure, strings.Join(messages, ",")))
		} else {
			results = append(results,
				internalobs.NewConditionFromPrefix(obsv1.ConditionValidOutputPrefix, out.Name, true, obsv1.ReasonValidationSuccess, fmt.Sprintf("output %q is valid", out.Name)))
		}
	}

	return common.AttributeConditionOutputs, results
}
