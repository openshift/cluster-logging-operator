package observability

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"

	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NameList provides a list of names for any resource implementing Names
type NameList interface {
	Names() []string
}

// DeployAsDeployment evaluates the spec to determine if the collector will be deployed as a deployment.
// Collector is not a daemonset if the only input source is an HTTP receiver
// Enabled through an annotation
func DeployAsDeployment(forwarder obs.ClusterLogForwarder) bool {
	if _, ok := forwarder.Annotations[constants.AnnotationEnableCollectorAsDeployment]; ok {
		inputTypes := Inputs(forwarder.Spec.Inputs).InputTypes()
		return len(inputTypes) == 1 && inputTypes[0] == obs.InputTypeReceiver
	}
	return false
}

// IsValidSpec evaluates the status conditions to determine if the spec is valid
func IsValidSpec(forwarder obs.ClusterLogForwarder) bool {
	log.V(3).Info("IsValidSpec", "outputs", forwarder.Spec.Outputs)
	status := forwarder.Status
	return isAuthorized(status.Conditions) &&
		isValid(obs.ConditionTypeValidInputPrefix, status.InputConditions, len(forwarder.Spec.Inputs)) &&
		isValid(obs.ConditionTypeValidOutputPrefix, status.OutputConditions, len(forwarder.Spec.Outputs)) &&
		isValid(obs.ConditionTypeValidPipelinePrefix, status.PipelineConditions, len(forwarder.Spec.Pipelines)) &&
		isValid(obs.ConditionTypeValidFilterPrefix, status.FilterConditions, len(forwarder.Spec.Filters))
}

func isValid(prefix string, conditions []metav1.Condition, expConditions int) bool {
	log.V(3).Info("isValid Args", "prefix", prefix, "conditions", conditions, "exp", expConditions)
	if len(conditions) != expConditions {
		return false
	}
	conditionTrue := 0
	for _, cond := range conditions {
		if strings.HasPrefix(cond.Type, prefix) && cond.Status == obs.ConditionTrue {
			conditionTrue++
		}
	}
	log.V(3).Info("isValid", "prefix", prefix, "act", conditionTrue, "exp", expConditions)
	return conditionTrue == expConditions
}

func isAuthorized(conditions []metav1.Condition) bool {
	for _, cond := range conditions {
		if cond.Type == obs.ConditionTypeAuthorized && cond.Status == obs.ConditionTrue {
			return true
		}
	}
	return false
}

type ClusterLogForwarderSpec obs.ClusterLogForwarderSpec

func (spec ClusterLogForwarderSpec) InputSpecsTo(out obs.OutputSpec) (results []obs.InputSpec) {
	inputs := Inputs(spec.Inputs).Map()
	found := map[string]obs.InputSpec{}
	for _, p := range spec.Pipelines {
		if sets.NewString(p.OutputRefs...).Has(out.Name) {
			for _, ref := range p.InputRefs {
				found[ref] = inputs[ref]
			}
		}
	}
	for _, input := range found {
		results = append(results, input)
	}
	return results
}
