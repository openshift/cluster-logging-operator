package matchers

import (
	//"fmt"
	"github.com/onsi/gomega/types"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/status"
	corev1 "k8s.io/api/core/v1"

	//"k8s.io/utils/diff"
	//"reflect"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// Match condition by type, status and reason if reason != "".
// Also match messageRegex if it is not empty.
func matchCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	var status corev1.ConditionStatus
	if s {
		status = corev1.ConditionTrue
	} else {
		status = corev1.ConditionFalse
	}
	fields := Fields{"Type": Equal(t), "Status": Equal(status)}
	if r != "" {
		fields["Reason"] = Equal(r)
	}
	if messageRegex != "" {
		fields["Message"] = MatchRegexp(messageRegex)
	}
	return MatchFields(IgnoreExtras, fields)
}

func HaveCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	return ContainElement(matchCondition(t, s, r, messageRegex))
}

func equalCondition(expected status.Condition) types.GomegaMatcher {
	fields := Fields{
		"Type":               Equal(expected.Type),
		"Status":             Equal(expected.Status),
		"Reason":             Equal(expected.Reason),
		"Message":            Equal(expected.Message),
		"LastTransitionTime": MatchLastTransitionTime(expected.LastTransitionTime),
	}
	return MatchFields(IgnoreExtras, fields)
}

// MatchConditions compares 2 conditions slices.
// For LastTransitionTime, 2 conditions are considered equal either if both timestamps are identical. Or, if
// expected[..].LastTransitionTime.Time is time 0 UTC, it will check if actual[..].LastTransitionTime has a time stamp
// within the last 5 minutes.
func MatchConditions(expected status.Conditions) types.GomegaMatcher {
	var conditionMatchers []interface{}
	for _, condition := range expected {
		conditionMatchers = append(conditionMatchers, equalCondition(condition))
	}
	return ConsistOf(conditionMatchers...)
}
