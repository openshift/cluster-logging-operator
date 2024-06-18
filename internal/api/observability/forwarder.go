package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

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

// IsValid evaluates the status conditions to determine if the spec is valid
func IsValid(forwarder obs.ClusterLogForwarder) bool {
	status := forwarder.Status
	return isAuthorized(status.Conditions) &&
		isValid(obs.ConditionTypeValidInputPrefix, status.Inputs, len(forwarder.Spec.Inputs)) &&
		isValid(obs.ConditionTypeValidOutputPrefix, status.Outputs, len(forwarder.Spec.Outputs)) &&
		isValid(obs.ConditionTypeValidPipelinePrefix, status.Pipelines, len(forwarder.Spec.Pipelines)) &&
		isValid(obs.ConditionTypeValidFilterPrefix, status.Filters, len(forwarder.Spec.Filters))
}

func isValid(prefix string, conditions []metav1.Condition, expConditions int) bool {
	if len(conditions) != expConditions {
		return false
	}
	conditionTrue := 0
	for _, cond := range conditions {
		if strings.HasPrefix(cond.Type, prefix) && cond.Status == obs.ConditionTrue {
			conditionTrue++
		}
	}
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
