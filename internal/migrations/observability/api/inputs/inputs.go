package inputs

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func ConvertInputs(loggingClfSpec *logging.ClusterLogForwarderSpec) []obs.InputSpec {
	obsSpecInputs := []obs.InputSpec{}

	for _, input := range loggingClfSpec.Inputs {
		obsInput := obs.InputSpec{
			Name: input.Name,
		}

		if input.Application != nil {
			obsInput.Type = obs.InputTypeApplication
			obsInput.Application = MapApplicationInput(input.Application)

		} else if input.Infrastructure != nil && input.Infrastructure.Sources != nil {
			obsInput.Type = obs.InputTypeInfrastructure
			obsInput.Infrastructure = MapInfrastructureInput(input.Infrastructure)

		} else if input.Audit != nil {
			obsInput.Type = obs.InputTypeAudit
			obsInput.Audit = MapAuditInput(input.Audit)

		} else if input.Receiver != nil {
			obsInput.Type = obs.InputTypeReceiver
			obsInput.Receiver = MapReceiverInput(input.Receiver)
		}

		obsSpecInputs = append(obsSpecInputs, obsInput)
	}
	return obsSpecInputs
}
