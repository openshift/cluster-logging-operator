package api

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

const (
	parseFilterName                = "converted-filter-parse"
	detectMultilineErrorFilterName = "converted-filter-detectMultilineError"
)

// convertFilters maps logging.Filters to observability.Filters
func convertFilters(loggingClfSpec *logging.ClusterLogForwarderSpec) []obs.FilterSpec {
	obsFilters := []obs.FilterSpec{}

	for _, filter := range loggingClfSpec.Filters {
		obsFilter := &obs.FilterSpec{
			Name: filter.Name,
		}

		switch filter.Type {
		case logging.FilterKubeAPIAudit:
			obsFilter.Type = obs.FilterTypeKubeAPIAudit
			if filter.KubeAPIAudit != nil {
				obsFilter.KubeAPIAudit = mapKubeApiAuditFilter(*filter.KubeAPIAudit)
			}
		case logging.FilterDrop:
			obsFilter.Type = obs.FilterTypeDrop
			if filter.DropTestsSpec != nil {
				obsFilter.DropTestsSpec = mapDropFilter(*filter.DropTestsSpec)
			}
		case logging.FilterPrune:
			obsFilter.Type = obs.FilterTypePrune
			if filter.PruneFilterSpec != nil {
				obsFilter.PruneFilterSpec = mapPruneFilter(*filter.PruneFilterSpec)
			}
		}

		obsFilters = append(obsFilters, *obsFilter)
	}
	return obsFilters
}

func mapKubeApiAuditFilter(loggingKubeApiAudit logging.KubeAPIAudit) *obs.KubeAPIAudit {
	return &obs.KubeAPIAudit{
		Rules:             loggingKubeApiAudit.Rules,
		OmitStages:        loggingKubeApiAudit.OmitStages,
		OmitResponseCodes: loggingKubeApiAudit.OmitResponseCodes,
	}
}

func mapDropFilter(loggingDropTest []logging.DropTest) []obs.DropTest {
	obsDropTests := []obs.DropTest{}
	for _, test := range loggingDropTest {
		obsDropConditions := []obs.DropCondition{}

		for _, cond := range test.DropConditions {
			obsDropConditions = append(obsDropConditions, obs.DropCondition{
				Field:      obs.FieldPath(cond.Field),
				Matches:    cond.Matches,
				NotMatches: cond.NotMatches,
			})
		}

		obsDropTests = append(obsDropTests, obs.DropTest{
			DropConditions: obsDropConditions,
		})
	}

	return obsDropTests
}

func mapPruneFilter(loggingPruneSpec logging.PruneFilterSpec) *obs.PruneFilterSpec {
	spec := &obs.PruneFilterSpec{}
	for _, in := range loggingPruneSpec.In {
		spec.In = append(spec.In, obs.FieldPath(in))
	}
	for _, notIn := range loggingPruneSpec.NotIn {
		spec.NotIn = append(spec.NotIn, obs.FieldPath(notIn))
	}

	return spec
}
