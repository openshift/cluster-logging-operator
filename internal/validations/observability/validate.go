package observability

import (
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/inputs"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/outputs"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	clfValidators = []func(internalcontext.ForwarderContext) (common.AttributeConditionType, []metav1.Condition){
		ValidatePermissions,
		inputs.Validate,
		outputs.Validate,
		pipelines.Validate,
	}
)

// ValidateClusterLogForwarder validates the forwarder spec that can not be accomplished using api attributes and returns a set of conditions that apply to the spec
func ValidateClusterLogForwarder(context internalcontext.ForwarderContext) map[common.AttributeConditionType][]metav1.Condition {
	conditionMap := map[common.AttributeConditionType][]metav1.Condition{}
	for _, validate := range clfValidators {
		conditionType, failures := validate(context)
		if conditions, found := conditionMap[conditionType]; found {
			conditionMap[conditionType] = append(conditions, failures...)
		} else {
			conditionMap[conditionType] = conditions
		}
	}
	return conditionMap
}
