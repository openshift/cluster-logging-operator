package observability

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclock "k8s.io/utils/clock"
)

// clock is used to set status condition timestamps.
// This variable makes it easier to test conditions.
var clock kubeclock.WithTickerAndDelayedExecution = &kubeclock.RealClock{}

func NewCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Time{Time: clock.Now()},
	}
}

func NewConditionFromPrefix(prefix, name string, conditionMet bool, reason, message string) metav1.Condition {
	conditionType := fmt.Sprintf("%s-%s", prefix, name)
	status := obs.ConditionTrue
	if !conditionMet {
		status = obs.ConditionFalse
	}
	return NewCondition(conditionType, status, reason, message)
}

// SetCondition adds (or updates) the set of conditions with the given
// condition. It returns a boolean value indicating whether the set condition
// is new or was a change to the existing condition with the same type.
func SetCondition(conditions *[]metav1.Condition, newCond metav1.Condition) bool {
	newCond.LastTransitionTime = metav1.Time{Time: clock.Now()}

	for i, condition := range *conditions {
		if condition.Type == newCond.Type {
			if condition.Status == newCond.Status {
				newCond.LastTransitionTime = condition.LastTransitionTime
			}
			changed := condition.Status != newCond.Status ||
				condition.Reason != newCond.Reason ||
				condition.Message != newCond.Message
			(*conditions)[i] = newCond
			return changed
		}
	}
	*conditions = append(*conditions, newCond)
	return true
}

// PruneConditions keeps only those conditions that match names from the spec
func PruneConditions(conditions *[]metav1.Condition, spec NameList, conditionTypePrefix string) {
	keepers := []metav1.Condition{}
	for _, condition := range *conditions {
		for _, name := range spec.Names() {
			if condition.Type == fmt.Sprintf("%s-%s", conditionTypePrefix, name) {
				keepers = append(keepers, condition)
			}
		}
	}
	*conditions = keepers
}
