package registry

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters/apiaudit"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters/viaq"
)

var (
	// registry is a registry of well known filter types and types
	// with reserved names that don't require spec (e.g. viaq)
	registry = map[string]func(*logging.FilterSpec) filters.Filter{
		viaq.Name:     viaq.New,
		apiaudit.Name: apiaudit.New,
	}
)

// LookupProto retrieves the filter used to generate a configuration element
func LookupProto(typeOrFilterName string, filterSpecs map[string]*logging.FilterSpec) filters.Filter {
	typeName := typeOrFilterName
	fSpec := filterSpecs[typeOrFilterName]
	if fSpec != nil {
		typeName = fSpec.Type
	}
	if f, found := registry[typeName]; found {
		return f(fSpec)
	}
	return nil
}
