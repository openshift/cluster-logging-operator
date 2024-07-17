package matchers

import (
	//"fmt"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/utils/diff"
	//"reflect"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// MatchCondition condition by type, status and reason if reason != "".
// Also match messageRegex if it is not empty.
func MatchCondition(conditionType string, conditionStatus bool, reason string, messageRegex string) types.GomegaMatcher {
	var status metav1.ConditionStatus
	if conditionStatus {
		status = metav1.ConditionTrue
	} else {
		status = metav1.ConditionFalse
	}
	fields := Fields{"Type": MatchRegexp(conditionType), "Status": Equal(status)}
	if reason != "" {
		fields["Reason"] = Equal(reason)
	}
	if messageRegex != "" {
		fields["Message"] = MatchRegexp(messageRegex)
	}
	return MatchFields(IgnoreExtras, fields)
}

func HaveCondition(reConditionType string, conditionTrue bool, reason string, messageRegex string) types.GomegaMatcher {
	return ContainElement(MatchCondition(reConditionType, conditionTrue, reason, messageRegex))
}

func equalCondition(expected metav1.Condition) types.GomegaMatcher {
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
func MatchConditions(expected []metav1.Condition) types.GomegaMatcher {
	var conditionMatchers []interface{}
	for _, condition := range expected {
		conditionMatchers = append(conditionMatchers, equalCondition(condition))
	}
	return ConsistOf(conditionMatchers...)
}
