package api

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertInputs(loggingClfSpec *logging.ClusterLogForwarderSpec) []obs.InputSpec {
	obsSpecInputs := []obs.InputSpec{}

	for _, input := range loggingClfSpec.Inputs {
		obsInput := obs.InputSpec{
			Name: input.Name,
		}

		if input.Application != nil {
			obsInput.Type = obs.InputTypeApplication
			obsInput.Application = mapApplicationInput(input.Application)

		} else if input.Infrastructure != nil && input.Infrastructure.Sources != nil {
			obsInput.Type = obs.InputTypeInfrastructure
			obsInput.Infrastructure = mapInfrastructureInput(input.Infrastructure)

		} else if input.Audit != nil {
			obsInput.Type = obs.InputTypeAudit
			obsInput.Audit = mapAuditInput(input.Audit)

		} else if input.Receiver != nil {
			obsInput.Type = obs.InputTypeReceiver
			obsInput.Receiver = mapReceiverInput(input.Receiver)
		}

		obsSpecInputs = append(obsSpecInputs, obsInput)
	}
	return obsSpecInputs
}

func mapNamespacedContainers(nsContainerSlice []logging.NamespaceContainerSpec) []obs.NamespaceContainerSpec {
	if nsContainerSlice == nil {
		return nil
	}
	obsNamespacedContainer := []obs.NamespaceContainerSpec{}

	for _, c := range nsContainerSlice {
		obsNamespacedContainer = append(obsNamespacedContainer, obs.NamespaceContainerSpec{
			Namespace: c.Namespace,
			Container: c.Container,
		})
	}
	return obsNamespacedContainer
}

func mapApplicationInput(loggingApp *logging.Application) *obs.Application {
	obsApp := &obs.Application{
		Selector: (*metav1.LabelSelector)(loggingApp.Selector),
	}

	obsApp.Includes = mapNamespacedContainers(loggingApp.Includes)
	obsApp.Excludes = mapNamespacedContainers(loggingApp.Excludes)

	if loggingApp.ContainerLimit != nil {
		obsApp.Tuning = &obs.ContainerInputTuningSpec{
			RateLimitPerContainer: (*obs.LimitSpec)(loggingApp.ContainerLimit),
		}
	}

	return obsApp
}

func mapInfrastructureInput(loggingInfra *logging.Infrastructure) *obs.Infrastructure {
	obsInfra := &obs.Infrastructure{
		Sources: []obs.InfrastructureSource{},
	}

	for _, source := range loggingInfra.Sources {
		obsInfra.Sources = append(obsInfra.Sources, obs.InfrastructureSource(source))
	}

	return obsInfra
}

func mapAuditInput(loggingAudit *logging.Audit) *obs.Audit {
	obsAudit := &obs.Audit{
		Sources: []obs.AuditSource{},
	}
	for _, source := range loggingAudit.Sources {
		obsAudit.Sources = append(obsAudit.Sources, obs.AuditSource(source))
	}
	return obsAudit
}

func mapReceiverInput(loggingReceiver *logging.ReceiverSpec) *obs.ReceiverSpec {
	obsReceiver := &obs.ReceiverSpec{}

	// HTTP Receiver
	if loggingReceiver.HTTP != nil {
		obsReceiver.Type = obs.ReceiverTypeHTTP
		obsReceiver.Port = loggingReceiver.HTTP.Port
		obsReceiver.HTTP = &obs.HTTPReceiver{
			Format: obs.HTTPReceiverFormat(loggingReceiver.HTTP.Format),
		}

		// Syslog Receiver
	} else if loggingReceiver.Syslog != nil {
		obsReceiver.Type = obs.ReceiverTypeSyslog
		obsReceiver.Port = loggingReceiver.Syslog.Port
	}
	return obsReceiver
}
