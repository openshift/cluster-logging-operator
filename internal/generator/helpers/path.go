package helpers

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
)

func ValuePath(resource *obs.ValueReference, formatter ...string) string {
	if resource == nil {
		return ""
	}
	if resource.SecretName != "" {
		return observability.SecretPath(resource.SecretName, resource.Key, formatter...)
	} else if resource.ConfigMapName != "" {
		return observability.ConfigPath(resource.ConfigMapName, resource.Key, formatter...)
	}
	return ""
}

func SecretPath(resource *obs.SecretReference, formatter ...string) string {
	if resource == nil || resource.SecretName == "" {
		return ""
	}
	return observability.SecretPath(resource.SecretName, resource.Key, formatter...)
}
