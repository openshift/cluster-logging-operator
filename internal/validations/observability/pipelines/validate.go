package pipelines

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"strings"
)

func Validate(context internalcontext.ForwarderContext) {
	inputs := internalobs.Inputs(context.Forwarder.Spec.Inputs).Map()
	outputs := internalobs.Outputs(context.Forwarder.Spec.Outputs).Map()
	filters := internalobs.FilterMap(context.Forwarder.Spec)
	var messages []string
	for _, pipelineSpec := range context.Forwarder.Spec.Pipelines {
		refMessages := validateRef(pipelineSpec, inputs, outputs, filters)
		if len(refMessages) > 0 {
			messages = append(messages, fmt.Sprintf("refs not found: %s", strings.Join(refMessages, ",")))
		}
		messages = append(messages, verifyHostNameNotFilteredForGCL(pipelineSpec, outputs, filters)...)
		if len(messages) > 0 {
			internalobs.SetCondition(&context.Forwarder.Status.PipelineConditions,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidPipelinePrefix, pipelineSpec.Name, false, obs.ReasonValidationFailure, strings.Join(messages, ",")))
		} else {
			internalobs.SetCondition(&context.Forwarder.Status.PipelineConditions,
				internalobs.NewConditionFromPrefix(obs.ConditionTypeValidPipelinePrefix, pipelineSpec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("pipeline %q is valid", pipelineSpec.Name)))
		}
	}

}

// validateRef validates the references defined for the pipeline actually reference a spec'd input,output, or filter
func validateRef(pipeline obs.PipelineSpec, inputs map[string]obs.InputSpec, outputs map[string]obs.OutputSpec, filters map[string]*obs.FilterSpec) (results []string) {

	var inputRefs []string
	for _, ref := range pipeline.InputRefs {
		if _, found := inputs[ref]; !found {
			inputRefs = append(inputRefs, ref)
		}
	}
	if len(inputRefs) > 0 {
		results = append(results, fmt.Sprintf("inputs%v", inputRefs))
	}

	var outputRefs []string
	for _, ref := range pipeline.OutputRefs {
		if _, found := outputs[ref]; !found {
			outputRefs = append(outputRefs, ref)
		}
	}
	if len(outputRefs) > 0 {
		results = append(results, fmt.Sprintf("outputs%v", outputRefs))
	}

	var filterRefs []string
	for _, ref := range pipeline.FilterRefs {
		if _, found := filters[ref]; !found {
			filterRefs = append(filterRefs, ref)
		}
	}
	if len(filterRefs) > 0 {
		results = append(results, fmt.Sprintf("filters%v", filterRefs))
	}
	return results
}

// verifyHostNameNotFilteredForGCL verifies that within a pipeline featuring a GCL sink and prune filters, the `.hostname` field is exempted from pruning.
func verifyHostNameNotFilteredForGCL(pipeline obs.PipelineSpec, outputs map[string]obs.OutputSpec, filters map[string]*obs.FilterSpec) (results []string) {
	if len(pipeline.FilterRefs) == 0 {
		return nil
	}

	for _, out := range pipeline.OutputRefs {
		if output, exists := outputs[out]; exists && output.Type == obs.OutputTypeGoogleCloudLogging {
			for _, f := range pipeline.FilterRefs {
				if filterSpec, ok := filters[f]; ok && prunesHostName(*filterSpec) {
					results = append(results, fmt.Sprintf("%q prunes the `.hostname` field which is required for output: %q of type %q.", filterSpec.Name, output.Name, output.Type))
				}
			}
		}
	}
	return results
}

// prunesHostName checks if a prune filter prunes the `.hostname` field
func prunesHostName(filter obs.FilterSpec) bool {
	if filter.Type != obs.FilterTypePrune {
		return false
	}

	hostName := ".hostname"

	inListPrunes := false
	notInListPrunes := false

	if filter.PruneFilterSpec.NotIn != nil {
		found := false
		for _, field := range filter.PruneFilterSpec.NotIn {
			if string(field) == hostName {
				found = true
				break
			}
		}
		if !found {
			inListPrunes = true
		}
	}

	if filter.PruneFilterSpec.In != nil {
		for _, field := range filter.PruneFilterSpec.In {
			if string(field) == hostName {
				notInListPrunes = true
				break
			}
		}
	}

	return inListPrunes || notInListPrunes
}
