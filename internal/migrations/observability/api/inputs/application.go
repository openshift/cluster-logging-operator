package inputs

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MapApplicationInput(loggingApp *logging.Application) *obs.Application {
	obsApp := &obs.Application{
		Selector: (*metav1.LabelSelector)(loggingApp.Selector),
	}

	obsApp.Includes = mapNamespacedContainers(loggingApp.Includes)
	obsApp.Excludes = mapNamespacedContainers(loggingApp.Excludes)

	// Convert the deprecated logging `namespaces` spec
	var obsNamespaces []obs.NamespaceContainerSpec
	if loggingApp.Namespaces != nil {
		obsNamespaces = mapNamespacesToNamespacedContainers(loggingApp.Namespaces)
	}

	// Combine with the includes list if applicable
	if obsApp.Includes != nil {
		obsApp.Includes = append(obsApp.Includes, obsNamespaces...)
	} else {
		obsApp.Includes = obsNamespaces
	}

	if loggingApp.ContainerLimit != nil {
		obsApp.Tuning = &obs.ContainerInputTuningSpec{
			RateLimitPerContainer: (*obs.LimitSpec)(loggingApp.ContainerLimit),
		}
	}

	return obsApp
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

func mapNamespacesToNamespacedContainers(namespaceList []string) []obs.NamespaceContainerSpec {
	obsNamespaces := []obs.NamespaceContainerSpec{}
	for _, namespace := range namespaceList {
		obsNamespaces = append(obsNamespaces, obs.NamespaceContainerSpec{Namespace: namespace})
	}
	return obsNamespaces
}
