package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "sigs.k8s.io/controller-runtime/pkg/client"
)

type AttributeConditionType string

const (
	// AttributeConditionConditions are conditions that apply generally to a ClusterLogForwarder
	AttributeConditionConditions AttributeConditionType = "conditions"

	// AttributeConditionInputs are conditions that apply to inputs
	AttributeConditionInputs AttributeConditionType = "inputs"

	// AttributeConditionFilters are conditions that apply to filters
	AttributeConditionFilters AttributeConditionType = "filters"

	// AttributeConditionPipelines are conditions that apply to pipelines
	AttributeConditionPipelines AttributeConditionType = "pipelines"

	// AttributeConditionOutputs are conditions that apply to outputs
	AttributeConditionOutputs AttributeConditionType = "outputs"
)

var (
	clfValidators = []func(client kubernetes.Client, spec obs.ClusterLogForwarderSpec) (AttributeConditionType, []metav1.Condition){
		validateInputs,
	}
)

// ValidateClusterLogForwarder validates the forwarder spec that can not be accomplished using api attributes and returns a set of conditions that apply to the spec
func ValidateClusterLogForwarder(client kubernetes.Client, spec obs.ClusterLogForwarderSpec) map[AttributeConditionType][]metav1.Condition {
	conditionMap := map[AttributeConditionType][]metav1.Condition{}
	for _, validate := range clfValidators {
		conditionType, failures := validate(client, spec)
		if conditions, found := conditionMap[conditionType]; found {
			conditionMap[conditionType] = append(conditions, failures...)
		} else {
			conditionMap[conditionType] = conditions
		}
	}
	return conditionMap
}

func validateInputs(client kubernetes.Client, spec obs.ClusterLogForwarderSpec) (AttributeConditionType, []metav1.Condition) {
	results := []metav1.Condition{}
	for _, i := range spec.Inputs {
		var conditions []metav1.Condition
		switch i.Type {
		case obs.InputTypeApplication:
			conditions = inputs.ValidateApplication(i)
		case obs.InputTypeInfrastructure:
			conditions = inputs.ValidateInfrastructure(i)
		case obs.InputTypeAudit:
			conditions = inputs.ValidateApplication(i)
		case obs.InputTypeReceiver:
			conditions = inputs.ValidateReceiver(i)
		}
		results = append(results, conditions...)
	}
	return AttributeConditionInputs, results
}
