package observability

import (
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
