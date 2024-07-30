package filters

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

const (
	ParseFilterName                = "converted-filter-parse"
	DetectMultilineErrorFilterName = "converted-filter-detect-multiline-error"
	OpenshiftLabelsFilterName      = "converted-openshift-labels"
)

// ConvertFilters maps logging.Filters to observability.Filters
func ConvertFilters(loggingClfSpec *logging.ClusterLogForwarderSpec) []obs.FilterSpec {
	obsFilters := []obs.FilterSpec{}

	for _, filter := range loggingClfSpec.Filters {
		obsFilter := &obs.FilterSpec{
			Name: filter.Name,
		}

		switch filter.Type {
		case logging.FilterKubeAPIAudit:
			obsFilter.Type = obs.FilterTypeKubeAPIAudit
			if filter.KubeAPIAudit != nil {
				obsFilter.KubeAPIAudit = MapKubeApiAuditFilter(*filter.KubeAPIAudit)
			}
		case logging.FilterDrop:
			obsFilter.Type = obs.FilterTypeDrop
			if filter.DropTestsSpec != nil {
				obsFilter.DropTestsSpec = MapDropFilter(*filter.DropTestsSpec)
			}
		case logging.FilterPrune:
			obsFilter.Type = obs.FilterTypePrune
			if filter.PruneFilterSpec != nil {
				obsFilter.PruneFilterSpec = MapPruneFilter(*filter.PruneFilterSpec)
			}
		}

		obsFilters = append(obsFilters, *obsFilter)
	}
	return obsFilters
}
