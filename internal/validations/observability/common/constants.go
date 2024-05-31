package common

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

	SecretsMap    = "secretMap"
	ConfigMapsMap = "configMapsMap"
)
