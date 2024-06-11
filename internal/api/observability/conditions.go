package observability

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
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
