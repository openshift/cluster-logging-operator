package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
	clfMigrations = []func(client kubernetes.Client, spec obs.ClusterLogForwarderSpec) (AttributeConditionType, []metav1.Condition){}
)

// ValidateClusterLogForwarder validates the forwarder spec that can not be accomplished using api attributes and returns a set of conditions that apply to the spec
func ValidateClusterLogForwarder(client kubernetes.Client, spec obs.ClusterLogForwarderSpec) map[AttributeConditionType][]metav1.Condition {
	conditionMap := map[AttributeConditionType][]metav1.Condition{}
	for _, validate := range clfMigrations {
		conditionType, failures := validate(client, spec)
		if conditions, found := conditionMap[conditionType]; found {
			conditionMap[conditionType] = append(conditions, failures...)
		} else {
			conditionMap[conditionType] = conditions
		}
	}
	return conditionMap
}
