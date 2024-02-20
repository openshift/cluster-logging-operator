package conditions

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/status"
	corev1 "k8s.io/api/core/v1"
)

func NotReady(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionReady, corev1.ConditionFalse, r, format, args...)
}

func CondInvalid(format string, args ...interface{}) status.Condition {
	return NotReady(logging.ReasonInvalid, format, args...)
}

func CondMissing(format string, args ...interface{}) status.Condition {
	return NotReady(logging.ReasonMissingResource, format, args...)
}

func CondDegraded(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionDegraded, corev1.ConditionTrue, r, format, args...)
}

func CondReadyWithMessage(r status.ConditionReason, format string, args ...interface{}) status.Condition {
	return logging.NewCondition(logging.ConditionReady, corev1.ConditionTrue, r, format, args...)
}

var CondReady = status.Condition{Type: logging.ConditionReady, Status: corev1.ConditionTrue}
