package api

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

func convertPipelines(logStoreSpec *logging.LogStoreSpec, loggingClfSpec *logging.ClusterLogForwarderSpec) ([]obs.PipelineSpec, []obs.FilterSpec, bool) {
	needDefaultOut := false
	createdFilters := sets.NewString()
	var pipelineFilterSpecs []obs.FilterSpec

	obsPipelines := []obs.PipelineSpec{}
	for _, pipeline := range loggingClfSpec.Pipelines {
		obsPipeline := &obs.PipelineSpec{
			Name:       pipeline.Name,
			InputRefs:  pipeline.InputRefs,
			FilterRefs: pipeline.FilterRefs,
		}

		if logStoreSpec != nil && referencesDefaultOutput(pipeline.OutputRefs) {
			needDefaultOut = true
			pipelineOutrefs := []string{}

			// keep non-default output refs
			for _, out := range pipeline.OutputRefs {
				if out != "default" {
					pipelineOutrefs = append(pipelineOutrefs, out)
				}
			}

			// Add default output name to output refs
			// Default name is `default-<LOGSTORE TYPE>`. E.g `default-elasticsearch`, `default-lokistack`
			pipelineOutrefs = append(pipelineOutrefs, DefaultName+string(logStoreSpec.Type))
			obsPipeline.OutputRefs = pipelineOutrefs
		} else {
			obsPipeline.OutputRefs = pipeline.OutputRefs
		}

		// Determine pipeline filters to create and add to filterRefs
		filters, filterRefs := generatePipelineFilters(pipeline, createdFilters)
		// Keep track of if detectmultiline and parse filters have been created
		for _, filter := range filters {
			if filter.Type == obs.FilterTypeDetectMultiline || filter.Type == obs.FilterTypeParse {
				createdFilters.Insert(string(filter.Type))
			}
		}

		// Add created filters to refs
		pipelineFilterSpecs = append(pipelineFilterSpecs, filters...)
		if len(filterRefs) > 0 {
			obsPipeline.FilterRefs = append(pipeline.FilterRefs, filterRefs...)
		}

		// Add converted pipelines to new observability pipelineSpec slice
		obsPipelines = append(obsPipelines, *obsPipeline)
	}

	return obsPipelines, pipelineFilterSpecs, needDefaultOut
}

func referencesDefaultOutput(outputs []string) bool {
	for _, out := range outputs {
		if out == "default" {
			return true
		}
	}
	return false
}

// generatePipelineFilters creates and adds pipeline filters previously set on the pipeline itself
func generatePipelineFilters(loggingPipelineSpec logging.PipelineSpec, createdFilters *sets.String) ([]obs.FilterSpec, []string) {
	obsPipelinefilterSpecs := []obs.FilterSpec{}
	addedFilterRefs := []string{}
	if loggingPipelineSpec.DetectMultilineErrors {
		addedFilterRefs = append(addedFilterRefs, detectMultilineErrorFilterName)
		if !createdFilters.Has(string(obs.FilterTypeDetectMultiline)) {
			filterDetectMultiline := &obs.FilterSpec{
				Type: obs.FilterTypeDetectMultiline,
				Name: detectMultilineErrorFilterName,
			}
			obsPipelinefilterSpecs = append(obsPipelinefilterSpecs, *filterDetectMultiline)
		}
	}
	if loggingPipelineSpec.Parse == "json" {
		addedFilterRefs = append(addedFilterRefs, parseFilterName)
		if !createdFilters.Has(string(obs.FilterTypeParse)) {
			filterParse := &obs.FilterSpec{
				Type: obs.FilterTypeParse,
				Name: parseFilterName,
			}
			obsPipelinefilterSpecs = append(obsPipelinefilterSpecs, *filterParse)
		}
	}
	if len(loggingPipelineSpec.Labels) != 0 {
		openshiftLabelFilterName := fmt.Sprintf("filter-%s-%s", loggingPipelineSpec.Name, openshiftLabelsFilterName)
		addedFilterRefs = append(addedFilterRefs, openshiftLabelFilterName)
		obsFilter := &obs.FilterSpec{
			Type:            obs.FilterTypeOpenshiftLabels,
			Name:            openshiftLabelFilterName,
			OpenShiftLabels: loggingPipelineSpec.Labels,
		}
		obsPipelinefilterSpecs = append(obsPipelinefilterSpecs, *obsFilter)
	}
	return obsPipelinefilterSpecs, addedFilterRefs
}
