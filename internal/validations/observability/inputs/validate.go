package inputs

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Validate(context internalcontext.ForwarderContext) (common.AttributeConditionType, []metav1.Condition) {
	results := []metav1.Condition{}
	for _, i := range context.Forwarder.Spec.Inputs {
		var conditions []metav1.Condition
		switch i.Type {
		case obs.InputTypeApplication:
			conditions = ValidateApplication(i)
		case obs.InputTypeInfrastructure:
			conditions = ValidateInfrastructure(i)
		case obs.InputTypeAudit:
			conditions = ValidateApplication(i)
		case obs.InputTypeReceiver:
			conditions = ValidateReceiver(i, context.Secrets, context.ConfigMaps)
		}
		results = append(results, conditions...)
	}
	return common.AttributeConditionInputs, results
}
