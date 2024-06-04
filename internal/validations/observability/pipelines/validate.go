package pipelines

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/validations/observability/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Validate(context internalcontext.ForwarderContext) (_ common.AttributeConditionType, results []metav1.Condition) {
	inputs := internalobs.Inputs(context.Forwarder.Spec.Inputs).Map()
	outputs := internalobs.Outputs(context.Forwarder.Spec.Outputs).Map()
	filters := internalobs.FilterMap(context.Forwarder.Spec)
	for _, pipelineSpec := range context.Forwarder.Spec.Pipelines {
		results = append(results, validateRef(pipelineSpec, inputs, outputs, filters)...)
		results = append(results, verifyHostNameNotFilteredForGCL(pipelineSpec, outputs, filters)...)
	}

	return common.AttributeConditionPipelines, results
}

// validateRef validates the references defined for the pipeline actually reference a spec'd input,output, or filter
func validateRef(pipeline obs.PipelineSpec, inputs map[string]obs.InputSpec, outputs map[string]obs.OutputSpec, filters map[string]*obs.FilterSpec) (results []metav1.Condition) {

	addCond := func(refType, ref, reason string) {
		results = append(results, internalobs.NewCondition(obs.ValidationCondition,
			metav1.ConditionTrue,
			reason,
			fmt.Sprintf(`pipeline %q references %s %q not found`, pipeline.Name, refType, ref),
		))
	}

	for _, ref := range pipeline.InputRefs {
		if _, found := inputs[ref]; !found {
			addCond("input", ref, obs.ReasonPipelineInputRefNotFound)
		}
	}
	for _, ref := range pipeline.OutputRefs {
		if _, found := outputs[ref]; !found {
			addCond("output", ref, obs.ReasonPipelineOutputRefNotFound)
		}
	}
	for _, ref := range pipeline.FilterRefs {
		if _, found := filters[ref]; !found {
			addCond("filter", ref, obs.ReasonPipelineFilterRefNotFound)
		}
	}
	return results
}

// verifyHostNameNotFilteredForGCL verifies that within a pipeline featuring a GCL sink and prune filters, the `.hostname` field is exempted from pruning.
func verifyHostNameNotFilteredForGCL(pipeline obs.PipelineSpec, outputs map[string]obs.OutputSpec, filters map[string]*obs.FilterSpec) (results []metav1.Condition) {
	if len(pipeline.FilterRefs) == 0 {
		return nil
	}

	for _, out := range pipeline.OutputRefs {
		if output, exists := outputs[out]; exists && output.Type == obs.OutputTypeGoogleCloudLogging {
			for _, f := range pipeline.FilterRefs {
				if filterSpec, ok := filters[f]; ok && prunesHostName(*filterSpec) {
					results = append(results, internalobs.NewCondition(obs.ValidationCondition,
						metav1.ConditionTrue,
						obs.ReasonFilterPruneHostname,
						fmt.Sprintf("%q prunes the `.hostname` field which is required for output: %q of type %q.", filterSpec.Name, output.Name, output.Type),
					))
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
			if field == hostName {
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
			if field == hostName {
				notInListPrunes = true
				break
			}
		}
	}

	return inListPrunes || notInListPrunes
}
