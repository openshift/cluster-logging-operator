package clusterlogforwarder

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	corev1 "k8s.io/api/core/v1"
)

var (

	// OutputDroppedCondition is the condition when an output is dropped from reconciliation
	OutputDroppedCondition status.ConditionType = "OutputDropped"

	// OutputNotReferencedReason is the reason for dropping an output when it is not referenced by a pipeline
	OutputNotReferencedReason status.ConditionReason = "OutputNotReferencedReason"
)

// DropUnreferencedOutputs removes unreferenced outputs from ClusterLogForwarder to support backwards compatibility with
// previous versions that handled this scenario gracefully
func DropUnreferencedOutputs(namespace, name string, spec loggingv1.ClusterLogForwarderSpec, logStore *loggingv1.LogStoreSpec, extras map[string]bool, logstoreSecretName, saTokenSecret string) (loggingv1.ClusterLogForwarderSpec, map[string]bool, []loggingv1.Condition) {
	warnings := []loggingv1.Condition{}
	outputRefs := sets.NewString()
	for _, p := range spec.Pipelines {
		outputRefs.Insert(p.OutputRefs...)
	}
	outputs := []loggingv1.OutputSpec{}
	for _, o := range spec.Outputs {
		if outputRefs.Has(o.Name) {
			outputs = append(outputs, o)
		} else {
			warnings = append(warnings, loggingv1.NewCondition(OutputDroppedCondition, corev1.ConditionTrue, OutputNotReferencedReason, "%s not referenced by any pipeline and not used during evaluation", o.Name))
		}
	}
	spec.Outputs = outputs
	return spec, extras, warnings
}
