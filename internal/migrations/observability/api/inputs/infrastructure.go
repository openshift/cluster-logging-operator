package inputs

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapInfrastructureInput(loggingInfra *logging.Infrastructure) *obs.Infrastructure {
	obsInfra := &obs.Infrastructure{
		Sources: []obs.InfrastructureSource{},
	}

	for _, source := range loggingInfra.Sources {
		obsInfra.Sources = append(obsInfra.Sources, obs.InfrastructureSource(source))
	}

	return obsInfra
}
